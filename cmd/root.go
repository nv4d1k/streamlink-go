/*
Copyright © 2023 Naumov Vadik <nv4d1k@ya.ru>
*/
package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/nv4d1k/streamlink-go/app/http/controllers"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var (
	isDebug       bool
	listenAddress string
	listenPort    int
	proxy         string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "streamlink-go",
	Short: "Live streaming forwarding service",
	/*Long: `A longer description that spans multiple lines and likely contains
	examples and usage of using your application. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,*/
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if !isDebug {
			gin.SetMode(gin.ReleaseMode)
		}
		r := gin.Default()
		r.Use(func(context *gin.Context) {
			context.Set("debug", isDebug)
			context.Next()
		})
		r.Use(func(ctx *gin.Context) {
			p := ctx.DefaultQuery("proxy", "")
			if proxy != "" {
				ctx.Set("proxy", proxy)
			}
			if p != "" {
				ctx.Set("proxy", p)
			}
		})
		r.GET("/:platform/:room", controllers.Forwarder)
		r.GET("/:platform/:room/*param", controllers.Forwarder)
		r.GET("/debug/pprof/", func(ctx *gin.Context) { pprof.Index(ctx.Writer, ctx.Request) })
		r.GET("/debug/pprof/:1", func(ctx *gin.Context) { pprof.Index(ctx.Writer, ctx.Request) })
		r.GET("/debug/pprof/trace", func(ctx *gin.Context) { pprof.Trace(ctx.Writer, ctx.Request) })
		r.GET("/debug/pprof/symbol", func(ctx *gin.Context) { pprof.Symbol(ctx.Writer, ctx.Request) })
		r.GET("/debug/pprof/profile", func(ctx *gin.Context) { pprof.Profile(ctx.Writer, ctx.Request) })
		r.GET("/debug/pprof/cmdline", func(ctx *gin.Context) { pprof.Cmdline(ctx.Writer, ctx.Request) })

		ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenAddress, listenPort))
		if err != nil {
			log.Fatalf("create listener error: %s\n", err.Error())
		}
		fmt.Printf("listening on %s ...\n", ln.Addr().String())
		fmt.Printf("access in player with room id. eg. http://%s/twitch/eslcs\n\n", ln.Addr().String())
		err = http.Serve(ln, r)
		if err != nil {
			log.Fatalf("http serve error: %s\n", err.Error())
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.streamlink-go.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolVar(&isDebug, "debug", false, "debug mode")
	rootCmd.PersistentFlags().StringVarP(&listenAddress, "listen-address", "l", "127.0.0.1", "listen address")
	rootCmd.PersistentFlags().IntVarP(&listenPort, "listen-port", "p", 8192, "listen port")
	rootCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "proxy url")
}
