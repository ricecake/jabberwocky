/*
Copyright © 2021 Sebastian Green-Husted <geoffcake@gmail.com>

*/
package cmd

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/apex/log"
	static "github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/olahol/melody.v1"

	"jabberwocky/cluster"
	"jabberwocky/storage"
	"jabberwocky/transport"
)

// content is our static web server content.
//go:embed content/*
var content embed.FS

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		notifyClose := make(chan os.Signal)
		signal.Notify(notifyClose, os.Interrupt)

		ctx, dbErr := storage.ConnectDb(ctx)
		if dbErr != nil {
			log.Fatal(dbErr.Error())
		}

		initErr := storage.InitTables(ctx)
		if initErr != nil {
			log.Fatal(initErr.Error())
		}

		unkErr := storage.MarkServersUnknown(ctx)
		if unkErr != nil {
			log.Fatal(unkErr.Error())
		}

		r := gin.Default()
		mAdmin := melody.New()
		mAgent := melody.New()
		r.Use(static.Serve("/", EmbedFolder(content, "content")))

		r.GET("/ws/admin", func(c *gin.Context) {
			log.Info("ws connection")
			mAdmin.HandleRequest(c.Writer, c.Request)
		})

		r.GET("/ws/agent", func(c *gin.Context) {
			log.Info("Client connection")
			mAgent.HandleRequest(c.Writer, c.Request)
		})

		mAgent.HandleConnect(func(s *melody.Session) {
			//This should broadcast a list of agents to the newly connected client, so that it can assess appropriately.
			log.Info("Websocket established")
			servers, err := storage.ListLiveServers(ctx)
			if err != nil {
				log.Error(err.Error())
			}

			msg, err := transport.Message{
				Type:    "serverList",
				Content: servers,
			}.EncodeJson()
			if err != nil {
				log.Error(err.Error())
			}
			s.Write(msg)
		})

		mAgent.HandleDisconnect(func(s *melody.Session) {
		})

		mAgent.HandleMessage(func(s *melody.Session, msg []byte) {
			log.Infof("Agent message: %#v", string(msg))
			body, err := transport.DecodeJson(msg)
			if err != nil {
				log.Error(err.Error())
				return
			}

			switch body.Type {
			// case "identity":
			// 	var ident transport.AgentIdentity
			// 	mapstructure.Decode(body.Content, &ident)
			// 	client = &storage.Client{
			// 		Uuid: ident.Uuid,
			// 		PublicKeyId: ident.PublicKeyId,
			// 		PublicKey: ident.PublicKey,
			// 	}
			// 	create client, save on session
			// 	if know the key, send challange
			// 	else, check validity of ident via configured mechanism
			// 		those configured mechanisms should be something like "assume they are who they say if we havent seen the key"
			// 		or "make an http call passing the info we know about the server".
			// 		for now, just implement the "trust" flow.
			// 		Challange response should probably be something like:
			// 			server sends random nonce.
			// 			agent hashes nonce with salt agent decides.
			// 			agent signs resulting hash, and sends hash and salt back to server
			// 			server hashes nonce with salt, and verifies signature.
			// 			Goal: agent never signs raw value chosen by another, but server can verify key ownership.
			// case "challangeResponse":
			// 	check the result of the challange against whats saved on session
			// case "setStatus":
			// 	do that
			// case "output":
			// 	forward to output handler
			default:
				mAdmin.Broadcast(msg)
			}
		})

		mAdmin.HandleMessage(func(s *melody.Session, msg []byte) {
			log.Info("got admin message")
			rep, err := transport.Message{
				Type: string(msg),
			}.EncodeJson()
			if err != nil {
				log.Error(err.Error())
			}
			mAgent.Broadcast(rep)
		})

		// Possible to make this use a dynamic self signed cert
		// by generating one with https://golang.org/pkg/crypto/x509/#MarshalPKCS8PrivateKey
		// and then using the RunTLS function.  Will want to gate behind a flag, as it's mostly useful for
		// testing.

		ginInterface := viper.GetString("server.http.interface")
		ginPort := viper.GetInt("server.http.port")
		ginRunOn := fmt.Sprintf("%s:%d", ginInterface, ginPort)

		if viper.GetBool("debug") {
			ginRunOn = ":0"
		}

		ln, err := net.Listen("tcp", ginRunOn)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Infof("Listening on %s", ln.Addr().String())

		nodeId, err := storage.GetNodeId(ctx)
		if err != nil {
			log.Fatal(err.Error())
		}

		go func() {
			http.Serve(ln, r)
		}()

		viper.SetDefault("server.advertise.host", ginInterface)
		viper.SetDefault("server.advertise.port", ginPort)

		err = storage.SaveServer(ctx, storage.Server{
			Uuid:   nodeId,
			Host:   viper.GetString("server.advertize.host"),
			Port:   viper.GetInt("server.advertize.port"),
			Status: "alive",
			Weight: 0,
		})

		if err != nil {
			log.Fatal(err.Error())
		}

		eventChan := make(chan cluster.MemberEvent, 1)

		err = cluster.StartCluster(ctx, eventChan)
		if err != nil {
			log.Fatal(err.Error())
		}

		for {
			select {
			case event := <-eventChan:
				log.Infof("NODE: %#v", event)
				err := storage.SaveServer(ctx, event.Server)
				if err != nil {
					log.Error(err.Error())
				}

				rep, err := transport.Message{
					Type:    "server",
					Content: event.Server,
				}.EncodeJson()
				if err != nil {
					log.Error(err.Error())
				}
				mAgent.Broadcast(rep)
			case <-notifyClose:
				cancel()
				log.Info("Shutting down per user request")
				time.Sleep(time.Duration(1) * time.Second)
				os.Exit(0)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	if err != nil {
		return false
	}
	return true
}

func EmbedFolder(fsEmbed embed.FS, targetPath string) static.ServeFileSystem {
	fsys, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}
