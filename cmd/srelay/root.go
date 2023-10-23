package srelay

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crelay",
	Short: "Cluster Relay proxy process for application",
	Long: `Cluster Relay process acts as a proxy for an application which doesnt want to receive active connections. 
			Instead, it receives the messages from the applications, and maintains a persistent connection`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
