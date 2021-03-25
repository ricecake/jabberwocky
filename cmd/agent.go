/*
Copyright © 2021 Sebastian Green-Husted <geoffcake@gmail.com>

*/
package cmd

import (
	"context"
	"fmt"

	"jabberwocky/payload"
	"jabberwocky/transport"

	"github.com/apex/log"
	"github.com/cenkalti/backoff/v4"
	"github.com/spf13/cobra"
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
		outerCtx, outerCancel := context.WithCancel(context.Background())
		defer outerCancel()
		ctxBackoff := backoff.WithContext(backoff.NewExponentialBackOff(), outerCtx)
		err := backoff.Retry(func() error {
			ctx, cancel := context.WithCancel(outerCtx)
			defer cancel()

			c, _, err := websocket.Dial(ctx, "wss://jabberwocky.devhost.dev/ws/agent", nil)
			if err != nil {
				log.Error(err.Error())
				return err
			}
			defer c.Close(websocket.StatusInternalError, "the sky is falling")

			errors := make(chan error)
			done := make(chan struct{})
			output := make(chan transport.Message)

			go func() {
				for msg := range output {
					err = wsjson.Write(ctx, c, msg)
					if err != nil {
						log.Error(err.Error())
						errors <- err
					}
				}
			}()

			input := make(chan transport.Message)
			go func() {
				for {
					var msg transport.Message
					err = wsjson.Read(ctx, c, &msg)
					if err != nil {
						log.Error(err.Error())
						errors <- err
					}
					input <- msg
				}
			}()

			for {
				select {
				case err := <-errors:
					return err
				case <-done:
					outerCancel()
				case msg := <-input:
					payload.Execute(ctx, msg, output)
				}
			}

			c.Close(websocket.StatusNormalClosure, "")
			return nil
		}, ctxBackoff)
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
