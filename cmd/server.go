/*
Copyright © 2021 Sebastian Green-Husted <geoffcake@gmail.com>

*/
package cmd

import (
"github.com/davecgh/go-spew/spew"
	"io/fs"
	"embed"
	"fmt"
	"net/http"
	"github.com/apex/log"
	static "github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gopkg.in/olahol/melody.v1"
)

// content is our static web server content.
//go:embed content/*
var content embed.FS


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
		a, b := content.ReadDir("content")
		fmt.Printf("%+v, %+v:", a, b)
		spew.Dump(a)
		spew.Dump(content)
		fmt.Println("server called")
		r := gin.Default()
		m := melody.New()
		r.Use(static.Serve("/", EmbedFolder(content, "content")))
//		r.StaticFS("/admin", http.FS(content))
//r.GET("/admin/*filepath", func(c *gin.Context) {
//	c.FileFromFS(c.Request.URL.Path, http.FS(content))
//})
		r.GET("/ws", func(c *gin.Context) {
			log.Info("ws connection")
			m.HandleRequest(c.Writer, c.Request)
		})
		m.HandleMessage(func(s *melody.Session, msg []byte) {
			m.Broadcast(msg)
		})

		r.Run(":5000")

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
