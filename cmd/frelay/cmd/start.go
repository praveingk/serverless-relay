package frelay

import (
	"fmt"
	"path/filepath"

	"github.com/clusterlink-net/clusterlink/pkg/util"
	"github.com/serverless-relay/cmd/config"
	relay "github.com/serverless-relay/pkg/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A start command starts the cluster relay",
	Long: `A start command set all parameter state of the MBg-
			The  id, IP cport(Cntrol port for grpc) and localDataPortRange,externalDataPortRange
			TBD now is done manually need to call some external `,
	Run: func(cmd *cobra.Command, args []string) {
		ip, _ := cmd.Flags().GetString("ip")
		port, _ := cmd.Flags().GetString("port")
		debug, _ := cmd.Flags().GetBool("debug")
		var rel relay.Relay
		ll := logrus.InfoLevel
		if debug == true {
			ll = logrus.DebugLevel
		}
		rel.Init(ip, port, ll)

		frelayDirectory := config.FrelayDirectory()

		// parse TLS files
		parsedCertData, err := util.ParseTLSFiles(config.FrCAFile,
			filepath.Join(frelayDirectory, config.CertificateFileName),
			filepath.Join(frelayDirectory, config.PrivateKeyFileName))
		if err != nil {
			fmt.Printf("Unable to parse TLS files: %v", err)
			return
		}

		rel.StartRelay(parsedCertData)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().String("ip", "", "Optional IP address to bind the serverless-relay")
	startCmd.Flags().String("port", "", "Port to bind the serverless-relay")
	startCmd.Flags().String("target", "", "Reachable IP:port or gateway service ID of the target service through Clusterlink gateway obtained through 'gwctl get import '")
	startCmd.Flags().Bool("debug", false, "Debug mode with verbose prints")
}
