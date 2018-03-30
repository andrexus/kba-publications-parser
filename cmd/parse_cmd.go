package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/andrexus/kba-publications-parser/conf"
	"github.com/andrexus/kba-publications-parser/service"
	"io/ioutil"
	"encoding/json"
)

var kbaParseCmd = cobra.Command{
	Use:   "parse",
	Short: "Parses KBA pdf files",
	Run: func(cmd *cobra.Command, args []string) {
		execWithConfig(cmd, parsePdf)
	},
}

func init() {
	kbaParseCmd.PersistentFlags().StringP("file", "f", "", "File to parse")
}

func parsePdf(cmd *cobra.Command, config *conf.Configuration) {
	fileToParse, err := cmd.Flags().GetString("file")
	if err != nil {
		logrus.Fatalf("%v", err)
	}
	logrus.Infof("Parsing KBA items from %s", fileToParse)

	rawData, err := ioutil.ReadFile(fileToParse)
	if err != nil {
		logrus.Fatalf("Can't read file. Error: %v", err)
	}

	kbaService := &service.KBAServiceImpl{}
	parseResults, err := kbaService.ParsePDF(rawData)
	if err != nil {
		logrus.Fatalf("Can't parse file. Error: %v", err)
	}
	jsonData, err := json.Marshal(parseResults)
	if err != nil {
		logrus.Fatalf("Can't create json file. Error: %v", err)
	}
	logrus.Info(string(jsonData))
}
