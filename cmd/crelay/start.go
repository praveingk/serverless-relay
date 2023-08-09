package crelay

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.ibm.com/cluster-relay/pkg/relay"
	api "github.ibm.com/mbg-agent/pkg/api"
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
		debug, _ := cmd.Flags().GetBool("debug")
		m := api.Mbgctl{Id: gw}
		var rel relay.Relay
		if !strings.Contains(target, ":") {
			// Check the target is a exported service
			sArr, err := m.GetRemoteService(target)
			if err != nil {
				fmt.Printf("Unable to get remote service : %+v", err)
			}
			target = sArr[0].Ip
		}
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
	startCmd.Flags().String("gw", "", "Name of the Clusterlink gateway control")
	startCmd.Flags().String("target", "", "Reachable IP:port or gateway service ID of the target service through Clusterlink gateway obtained through 'gwctl get service'")
	startCmd.Flags().Bool("debug", false, "Debug mode with verbose prints")
}
