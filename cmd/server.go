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
	"path"
	"strings"
	"time"

	"github.com/apex/log"
	static "github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/ricecake/karma_chameleon/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/olahol/melody.v1"

	"jabberwocky/cluster"
	"jabberwocky/processing"
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

		nodeId, err := storage.GetNodeId(ctx)
		if err != nil {
			log.Fatal(err.Error())
		}

		markServErr := storage.MarkServersUnknown(ctx)
		if markServErr != nil {
			log.Fatal(markServErr.Error())
		}

		markAgentErr := storage.MarkAgentsUnknown(ctx)
		if markAgentErr != nil {
			log.Fatal(markAgentErr.Error())
		}

		r := gin.Default()
		mClient := melody.New()
		mAgent := melody.New()
		r.Use(static.Serve("/", EmbedFolder(content, "content")))

		r.GET("/ws/admin", func(c *gin.Context) {
			mClient.HandleRequest(c.Writer, c.Request)
		})

		r.GET("/ws/agent", func(c *gin.Context) {
			code := c.Request.Header.Get("Agent-Id")
			if code == "" {
				c.Status(403)
				return
			}
			//HandleRequestWithKeys
			// Can use that to extract auth information from headers, then validate it and populate the Agent object on the session from the get go.
			mAgent.HandleRequestWithKeys(c.Writer, c.Request, map[string]interface{}{
				"code": code,
			})
		})

		mClient.HandleConnect(func(s *melody.Session) {
			code := util.CompactUUID()
			log.Infof("Admin Connected %s", code)
			s.Set("code", code)
			// this will also want to register something that says that messages for this code go to this websocket.
			// probably by setting it up in the cluster router.
			// will need to figure out how to handle closing the channel, and making sure that happens when the websocket closes.
			// maybe have a function that asks the router to close the channel, and then the websocket listenr listens for channel close, and shuts down the goroutrine? <- yes.
			channel := cluster.Router.RegisterClient(code)
			go func() {
				for {
					select {
					case msg, more := <-channel:
						if !more {
							log.Infof("Closing %s", code)
							if !s.IsClosed() {
								s.Close()
							}
							return
						}

						msgRep, err := msg.EncodeJson()
						if err != nil {
							log.Error(err.Error())
							continue
						}
						s.Write(msgRep)
					}
				}
			}()

			cluster.Router.Emit(cluster.LOCAL_CLIENT, transport.NewMessage("client", "connect", code))
		})

		mClient.HandleDisconnect(func(s *melody.Session) {
			code := s.MustGet("code").(string)
			cluster.Router.UnregisterClient(code)
			cluster.Router.Emit(cluster.LOCAL_CLIENT, transport.NewMessage("client", "disconnect", code))
		})

		mClient.HandleMessage(func(s *melody.Session, msg []byte) {
			code := s.MustGet("code").(string)

			encMsg, err := transport.DecodeJson(msg)
			if err != nil {
				log.Error(err.Error())
			}

			encMsg.SourceId = code

			cluster.Router.Emit(cluster.LOCAL_CLIENT, encMsg)
		})

		mAgent.HandleConnect(func(s *melody.Session) {
			code := s.MustGet("code").(string)
			log.Infof("Agent Connected %s", code)
			channel := cluster.Router.RegisterAgent(code)
			go func() {
				for {
					select {
					case msg, more := <-channel:
						if !more {
							log.Infof("Closing routing %s", code)
							if !s.IsClosed() {
								s.Close()
							}
							return
						}

						msgRep, err := msg.EncodeJson()
						if err != nil {
							log.Error(err.Error())
							continue
						}
						s.Write(msgRep)
					}
				}
			}()

			cluster.Router.Emit(cluster.LOCAL_AGENT, transport.NewMessage("agent", "connect", storage.Agent{
				Uuid:            code,
				DelegatedServer: nodeId,
				Status:          "connected",
				LastContact:     time.Now(),
			}))
		})

		mAgent.HandleDisconnect(func(s *melody.Session) {
			code := s.MustGet("code").(string)
			log.Infof("Closing websocket: %s", code)
			cluster.Router.UnregisterAgent(code)
			cluster.Router.Emit(cluster.LOCAL_AGENT, transport.NewMessage("agent", "disconnect", storage.Agent{
				Uuid:            code,
				DelegatedServer: nodeId,
				Status:          "disconnected",
				LastContact:     time.Now(),
			}))

		})

		mAgent.HandleMessage(func(s *melody.Session, msg []byte) {
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
				cluster.Router.Emit(cluster.LOCAL_AGENT, body)
			}
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

		go func() {
			http.Serve(ln, r)
		}()

		viper.SetDefault("server.advertise.host", ginInterface)
		viper.SetDefault("server.advertise.port", ginPort)

		err = storage.SaveServer(ctx, storage.Server{
			Uuid:   nodeId,
			Host:   viper.GetString("server.advertise.host"),
			Port:   viper.GetInt("server.advertise.port"),
			Status: "alive",
			Weight: 0,
		})

		if err != nil {
			log.Fatal(err.Error())
		}

		err = cluster.StartCluster(ctx)
		if err != nil {
			log.Fatal(err.Error())
		}

		go processing.HandleStorage(ctx)
		go processing.HandleOutput(ctx)

		for {
			select {
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
}

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, reqPath string) bool {
	if reqPath != "/" {
		reqPath = strings.TrimSuffix(reqPath, "/")
	}

	file, err := e.Open(reqPath)
	if err != nil {
		return false
	}

	stats, err := file.Stat()
	if err != nil {
		return false
	}

	if stats.IsDir() {
		index := path.Join(reqPath, "index.html")
		_, err := e.Open(index)
		if err != nil {
			return false
		}
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
