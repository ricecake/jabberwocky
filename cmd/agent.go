/*
Copyright © 2021 Sebastian Green-Husted <geoffcake@gmail.com>

*/
package cmd

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"jabberwocky/payload"
	"jabberwocky/storage"
	"jabberwocky/transport"
	"jabberwocky/util"

	"github.com/apex/log"
	"github.com/cenkalti/backoff/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "A brief description of your command",
	Long: `A lo	nger description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("agent called")
		/*
			setup storage libs, or load db from disk
			make dns checks to find list to join, if available
			check database for server list, if available
			check config for main node
			in backoff loop, try to connect to best fit.
			if can't get connected, try the next best fit.
			proceed through loop forver, with backoff doing it's thing.
			repeat if options expanded.
			once connected, execute
				boot jobs
				on connect jobs
				start timers and crons
		*/

		// This context backoff doesn't do what was wanted.  Might be useless to me.
		initCtx, outerCancel := context.WithCancel(context.Background())
		defer outerCancel()

		outerCtx, dbErr := storage.ConnectDb(initCtx)
		if dbErr != nil {
			log.Fatal(dbErr.Error())
		}

		storage.InitTables(outerCtx)

		output := make(chan transport.Message)
		input := make(chan transport.Message)

		go func() {
		CBLOOP:
			for {
				select {
				case <-outerCtx.Done():
					break CBLOOP
				case msg := <-input:
					payload.Execute(outerCtx, msg, output)
				}
			}
		}()

		ctxBackoff := backoff.WithContext(backoff.NewExponentialBackOff(), outerCtx)
		err := backoff.RetryNotify(func() error {
			errors := make(chan error)
			ctx, cancel := context.WithCancel(outerCtx)
			defer cancel()

			//TODO: use rendezvous hashing to pick server
			//      need to make sure that our "seed node" from the config/dns/wherever is in there, since db only has "seen" nodes from cluster.
			//      Might be easiest to just get list from db, and if empty, populate with defaults.  That way we might only connect to seed node once.

			agentId, err := storage.GetNodeId(ctx)
			if err != nil {
				return err
			}

			var servers []string
			seen, err := storage.ListLiveServers(ctx)
			if err != nil {
				return err
			}

			if len(seen) == 0 {
				log.Info("Using seed nodes")
				servers = viper.GetStringSlice("agent.seed_nodes")
			} else {
				for _, serv := range seen {
					servers = append(servers, serv.UrlString())
				}
			}

			hrw := util.NewHrw()

			hrw.AddNode(servers...)

			targetNode := hrw.Get(agentId)
			log.Infof("picked server [%s]", targetNode)

			connUrl, err := url.Parse(targetNode)
			if err != nil {
				return err
			}

			connUrl.Scheme = "wss"
			connUrl.Path = "/ws/agent"

			c, _, err := websocket.Dial(ctx, connUrl.String(), nil)
			if err != nil {
				return err
			}
			defer c.Close(websocket.StatusInternalError, "Unexpected disconnection")
			log.Info("Connected to server")

			go func() {
			output:
				for {
					select {
					case <-ctx.Done():
						break output
					case msg := <-output:
						/*
							There should we a handler here that will check for the type of the message, and if it's a control message,
							then it should do the right control action.
							The specific desired course is that the payload handler decides that it needs to re-do the connection,
							so it sends a control message saying to reconnect.
							The message will get sent to the server cluster, and then it'll disconnect, and start the reconnect flow.
							This precludes moving a lot of the "not websocket" logic out of the backoff loop, which should make it behave easier
							in the desired direction.
						*/
						err = wsjson.Write(ctx, c, msg)
						if err != nil {
							errors <- err
						}
						switch msg.Type {
						case "shutdown":
							outerCancel()
						case "reconnect":
							cancel()
						}
					}
				}
			}()

			go func() {
				for {
					var msg transport.Message
					err = wsjson.Read(ctx, c, &msg)
					if err != nil {
						errors <- err
					}
					input <- msg
				}
			}()

			select {
			case err := <-errors:
				return err
			case <-ctx.Done():
				c.Close(websocket.StatusNormalClosure, "Planned shutdown")
			}

			return nil
		}, ctxBackoff, func(err error, backoff time.Duration) {
			log.Errorf("Backoff [%0.2f]: %s", backoff.Seconds(), err.Error())
		})

		if err != nil {
			log.Fatal(err.Error())
		}

	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// agentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// agentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
