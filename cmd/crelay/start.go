package crelay

import (
	"github.com/clusterlink-host-relay/pkg/relay"
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
		target, _ := cmd.Flags().GetString("target")
		debug, _ := cmd.Flags().GetBool("debug")
		var rel relay.Relay
		ll := logrus.InfoLevel
		if debug == true {
			ll = logrus.DebugLevel
		}
		rel.Init(ip, port, target, ll)
		rel.StartRelay()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().String("ip", "", "Optional IP address to bind the cluster-relay")
	startCmd.Flags().String("port", "", "Port to bind the cluster-relay")
	startCmd.Flags().String("target", "", "Reachable IP:port or gateway service ID of the target service through Clusterlink gateway obtained through 'gwctl get import '")
	startCmd.Flags().Bool("debug", false, "Debug mode with verbose prints")
}
