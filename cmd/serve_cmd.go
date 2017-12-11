package cmd

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xlab/closer"

	"github.com/andrexus/kba-publications-parser/conf"
	"github.com/andrexus/kba-publications-parser/api"
)

var serveCmd = cobra.Command{
	Use:   "serve",
	Short: "Start API server",
	Long:  "Start API server on specified host and port",
	Run: func(cmd *cobra.Command, args []string) {
		execWithConfig(cmd, serve)
	},
}

func serve(config *conf.Configuration) {

	apiServer := api.NewAPI(config)

	closer.Bind(func() {
		err := apiServer.Stop()
		if err != nil {
			logrus.Errorf("Error stopping API server: %s", err.Error())
		}
	})

	l := fmt.Sprintf("%v:%v", config.API.Host, config.API.Port)
	logrus.Infof("API started on: %s", l)
	logrus.Fatal(apiServer.Start())
}
