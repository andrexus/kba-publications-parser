package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/andrexus/kba-publications-parser/conf"
)

var rootCmd = cobra.Command{
	Use:  "kba-publications-parser",
	Long: "A service that will start API by default.",
	Run: func(cmd *cobra.Command, args []string) {
		execWithConfig(cmd, serve)
	},
}

// NewRoot will add flags and subcommands to the different commands
func RootCmd() *cobra.Command {
	rootCmd.PersistentFlags().StringP("config", "c", "", "The configuration file")
	rootCmd.AddCommand(&serveCmd, &kbaParseCmd, &versionCmd)
	return &rootCmd
}

func execWithConfig(cmd *cobra.Command, fn func(cmd *cobra.Command, config *conf.Configuration)) {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		logrus.Fatalf("%+v", err)
	}

	config, err := conf.Load(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %+v", err)
	}

	fn(cmd, config)
}
