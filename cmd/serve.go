/*
Copyright © 2023 Naumov Vadik <nv4d1k@ya.ru>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	gServeListenAddress string
	gServeListenPort    int
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start serving mode",
	/*Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("serve called")
		},*/
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&gServeListenAddress, "listen-address", "a", "127.0.0.1", "Listening address")
	serveCmd.PersistentFlags().IntVarP(&gServeListenPort, "listen-port", "p", 0, "Listening port")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
