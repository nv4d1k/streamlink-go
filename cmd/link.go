/*
Copyright © 2023 Naumov Vadik <nv4d1k@ya.ru>
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/nv4d1k/streamlink-go/streamlibs/HuYa"
	"github.com/nv4d1k/streamlink-go/streamlibs/Twitch"
	"github.com/spf13/cobra"
)

var (
	gLinkSite string
)

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Get link of stream",
	/*Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,*/
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatalln("Room ID was not supplied!")
		}
		switch strings.ToLower(gLinkSite) {
		case "huya":
			hy, err := HuYa.NewHuyaLink(args[0], gProxyURL)
			if err != nil {
				log.Fatalln(err.Error())
			}
			link, err := hy.GetLink()
			if err != nil {
				log.Fatalln(err.Error())
			}
			fmt.Println(link)
		case "twitch":
			tw, err := Twitch.NewTwitchLink(args[0], gProxyURL)
			if err != nil {
				log.Fatalln(err.Error())
			}
			link, err := tw.GetLink()
			if err != nil {
				log.Fatalln(err.Error())
			}
			fmt.Println(link)
		default:
			fmt.Println("Unsupported site.")
		}
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// linkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// linkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	linkCmd.PersistentFlags().StringVarP(&gLinkSite, "site", "s", "", "Stream web site")
}
