package admin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/clusterlink-net/clusterlink/cmd/cl-adm/util"
	"github.com/spf13/cobra"

	"github.com/serverless-relay/cmd/config"
)

func signCerts(entity string) {
	partyDirectory := config.PartyDirectory(entity)

	util.CreateCertificate(&util.CertificateConfig{
		Name:              entity,
		IsClient:          true,
		CAPath:            filepath.Join(partyDirectory, config.CertificateFileName),
		CAKeyPath:         filepath.Join(partyDirectory, config.PrivateKeyFileName),
		CertOutPath:       filepath.Join(partyDirectory, config.CertificateFileName),
		PrivateKeyOutPath: filepath.Join(partyDirectory, config.PrivateKeyFileName),
	})

}
func createRelay() {
	fmt.Printf("Creating Frelay CA Cert.")
	if err := os.MkdirAll(config.BaseDirectory(), 0755); err != nil {
		fmt.Printf("Unable to create directory :%v\n", err)
		return
	}
	err := util.CreateCertificate(&util.CertificateConfig{
		Name:              config.FrelayServerName,
		IsCA:              true,
		CertOutPath:       config.FrCAFile,
		PrivateKeyOutPath: config.FrKeyFile,
	})
	if err != nil {
		fmt.Printf("Unable to generate CA certficate :%v\n", err)
		return
	}
	fmt.Printf("Generating Certs/Key using CA.\n")

	frelayDirectory := config.FrelayDirectory()
	if err := os.MkdirAll(frelayDirectory, 0755); err != nil {
		fmt.Printf("Unable to create directory :%v\n", err)
		return
	}
	err = util.CreateCertificate(&util.CertificateConfig{
		Name:              config.FrelayServerName,
		IsServer:          true,
		IsClient:          true,
		DNSNames:          []string{config.FrelayServerName},
		CAPath:            config.FrCAFile,
		CAKeyPath:         config.FrKeyFile,
		CertOutPath:       filepath.Join(frelayDirectory, config.CertificateFileName),
		PrivateKeyOutPath: filepath.Join(frelayDirectory, config.PrivateKeyFileName),
	})
	if err != nil {
		fmt.Printf("Unable to generate certficate/key :%v\n", err)
		return
	}
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a relay/party",
	Long:  `Create a relay/party `,
	Run: func(cmd *cobra.Command, args []string) {
		createRelay()
	},
}

var createRelayCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a relay",
	Long:  `Create a relay `,
	Run: func(cmd *cobra.Command, args []string) {
		createRelay()
	},
}

var createPartyCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a party",
	Long:  `Create a party `,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		signCerts(name)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createRelayCmd)
	createCmd.AddCommand(createPartyCmd)
	createPartyCmd.Flags().String("name", "", "Party name.")

}
