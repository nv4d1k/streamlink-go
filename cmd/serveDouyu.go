/*
Copyright © 2023 Naumov Vadik <nv4d1k@ya.ru>
*/
package cmd

import (
	"fmt"
	"github.com/nv4d1k/streamlink-go/streamlibs/DouYu"

	"github.com/spf13/cobra"
)

// serveDouyuCmd represents the serveDouyu command
var serveDouyuCmd = &cobra.Command{
	Use:   "douyu",
	Short: "Douyu",
	/*Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,*/
	Run: func(cmd *cobra.Command, args []string) {
		srv := DouYu.NewServer(fmt.Sprintf("%s:%d", gServeListenAddress, gServeListenPort), gProxyURL, gDebug)
		srv.Play()
		<-(make(chan struct{}))
	},
}

func init() {
	serveCmd.AddCommand(serveDouyuCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveDouyuCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveDouyuCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
