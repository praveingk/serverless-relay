package config

import "path/filepath"

const (
	//FrelayServerName is the servername used in frelay router
	FrelayServerName = "frelay"
	//CertsRootDirectory is the directory for storing all certs it can be changed to /etc/ssl/certs
	CertsRootDirectory = "certs"

	// FrCAFile is the path to the certificate authority file.
	FrCAFile = CertsRootDirectory + "/frelay-ca.pem"
	// FrKeyFile is the path to the private-key file.
	FrKeyFile = CertsRootDirectory + "/frelay-key.pem"

	// PrivateKeyFileName is the filename used by private key files.
	PrivateKeyFileName = "key.pem"
	// CertificateFileName is the filename used by certificate files.
	CertificateFileName = "cert.pem"
)

// BaseDirectory returns the base path of the fabric.
func BaseDirectory() string {
	return CertsRootDirectory
}

// PartyDirectory returns the base path for a specific party.
func PartyDirectory(party string) string {
	return filepath.Join(BaseDirectory(), party)
}

// PartyDirectory returns the base path for a Frelay
func FrelayDirectory() string {
	return filepath.Join(BaseDirectory(), FrelayServerName)
}
