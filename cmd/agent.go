/*
Copyright © 2021 Sebastian Green-Husted <geoffcake@gmail.com>

*/
package cmd

import (
	"context"
	"fmt"
	"net/http"
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

		// TODO
		// This needs to be setup so that it puts the messages into either a queue, or into a channel to be sent over the websocket.
		// The goal is that all the log messages will end up being shipped into the cluster.

		// log.SetHandler(log.HandlerFunc(func(e *log.Entry) error {
		// 	sendIt := time.Now().After(onlyAfter)
		// 	if sendIt {
		// 		_, anounceErr := discord.ChannelMessageSend(viper.GetString("channel"), fmt.Sprintf("%s: %s", e.Level, e.Message))
		// 		if anounceErr != nil {
		// 			return anounceErr
		// 		}
		// 	}
		// 	return nil
		// }))

		// This context backoff doesn't do what was wanted.  Might be useless to me.
		initCtx, outerCancel := context.WithCancel(context.Background())
		defer outerCancel()

		outerCtx, dbErr := storage.ConnectDb(initCtx)
		if dbErr != nil {
			log.Fatal(dbErr.Error())
		}

		storage.InitTables(outerCtx)
		storage.MarkServersUnknown(outerCtx)

		output := make(chan transport.Message)
		input := make(chan transport.Message)

		go func() {
			for {
				select {
				case <-outerCtx.Done():
					return
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

			agentId, err := storage.GetNodeId(ctx)
			if err != nil {
				return err
			}

			var servers []util.HrwNode

			knownLive, err := storage.ListLiveServers(ctx)
			if err != nil {
				return err
			}

			knownUnknown, err := storage.ListUnknownServers(ctx)
			if err != nil {
				return err
			}

			knownAny, err := storage.ListAllServers(ctx)
			if err != nil {
				return err
			}

			if len(knownLive) != 0 {
				for _, serv := range knownLive {
					servers = append(servers, serv)
				}
			} else if len(knownUnknown) != 0 {
				for _, serv := range knownUnknown {
					servers = append(servers, serv)
				}
			} else if len(knownAny) != 0 {
				storage.MarkServersUnknown(ctx)
				for _, serv := range knownAny {
					servers = append(servers, serv)
				}
			} else {
				log.Info("Using seed nodes")
			SEED:
				for _, sUrl := range viper.GetStringSlice("agent.seed_nodes") {
					serv, err := storage.ServerFromString(sUrl)
					if err != nil {
						log.Error(err.Error())
						continue SEED
					}
					servers = append(servers, serv)
				}
			}

			log.Infof("Using servers: %+v", servers)

			hrw := util.NewHrw()

			hrw.AddNode(servers...)

			targetNode := hrw.Get(agentId).(storage.Server)
			connUrl := targetNode.Url()

			connUrl.Scheme = "wss"
			connUrl.Path = "/ws/agent"

			log.Infof("picked server [%s]", connUrl.String())

			// This should include some fancy headers, to indicate the id of the agent, which can be pulled off by the server.
			c, _, err := websocket.Dial(ctx, connUrl.String(), &websocket.DialOptions{
				HTTPHeader: http.Header{
					"Agent-Id": []string{agentId},
				},
			})
			if err != nil {
				targetNode.Status = "degraded"
				storage.SaveServer(ctx, targetNode)
				return err
			}
			defer c.Close(websocket.StatusInternalError, "Unexpected disconnection")
			storage.SetCurrentServer(ctx, targetNode)
			log.Info("Connected to server")

			go func() {
				for {
					select {
					case <-ctx.Done():
						return
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
						switch msg.Type {
						case "control":
							switch msg.SubType {
							case "shutdown":
								outerCancel()
							case "reconnect":
								errors <- fmt.Errorf("Reconnection")
							}
						default:
							err = wsjson.Write(ctx, c, msg)
							if err != nil {
								errors <- err
							}
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
						return
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

		log.Info("Exiting")

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
