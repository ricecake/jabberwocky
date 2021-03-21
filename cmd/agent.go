/*
Copyright © 2021 Sebastian Green-Husted <geoffcake@gmail.com>

*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"jabberwocky/payload"
	"jabberwocky/transport"

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

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c, _, err := websocket.Dial(ctx, "wss://jabberwocky.devhost.dev/ws/agent", nil)
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		done := make(chan struct{})
		output := make(chan transport.Message)

		go func() {
			for msg := range output {
				err = wsjson.Write(ctx, c, msg)
				if err != nil {
					log.Fatal(err)
				}
			}
		}()

		input := make(chan transport.Message)
		go func() {
			for {
				var msg transport.Message
				err = wsjson.Read(ctx, c, &msg)
				if err != nil {
					log.Fatal(err)
				}
				input <- msg
			}
		}()

		for {
			select {
			case <-done:
				cancel()
			case msg := <-input:
				payload.Execute(ctx, msg, output)
			}
		}

		c.Close(websocket.StatusNormalClosure, "")

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
