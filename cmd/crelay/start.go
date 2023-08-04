package crelay

import (
	"github.com/spf13/cobra"
	"github.ibm.com/cluster-relay/pkg/relay"
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
		gw, _ := cmd.Flags().GetString("gw")
		target, _ := cmd.Flags().GetString("target")

		var rel relay.Relay
		rel.Init(ip, port, gw, target)
		rel.StartRelay()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().String("ip", "", "Optional IP address to bind the cluster-relay")
	startCmd.Flags().String("port", "", "Port to bind the cluster-relay")
	startCmd.Flags().String("gw", "", "Reachable IP of the  Clusterlink gateway")
	startCmd.Flags().String("target", "", "Reachable IP:port of the target service through Clusterlink gateway obtained through 'gwctl get service'")
}
