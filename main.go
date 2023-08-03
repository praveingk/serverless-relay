/* Cluster-Relay :
 * 1) Input : destination
 */

package main

import (
	"os"

	"githib.ibm.com/cluster-relay/pkg/relay"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "Cluster Relay",
	Short: "Cluster Relay proxy process for application",
	Long: `Cluster Relay process acts as a proxy for an application which doesnt want to receive active connections. 
			Instead, it receives the messages from the applications, and maintains a persistent connection`,
}

// / startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A start command set all parameter state of the Multi-cloud Border Gateway",
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

func executeClusterRelay() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().String("ip", "", "Optional IP address to bind the cluster-relay")
	startCmd.Flags().String("port", "", "Port to bind the cluster-relay")
	startCmd.Flags().String("gw", "", "Reachable IP of the  Clusterlink gateway")
	startCmd.Flags().String("target", "", "Reachable IP:port of the target service through Clusterlink gateway obtained through 'gwctl get service'")

}

func main() {
	executeClusterRelay()
}
