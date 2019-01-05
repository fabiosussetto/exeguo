package cmd

import (
	"fmt"

	"github.com/fabiosussetto/exeguo/security"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// func init() {
// 	// ServerCmd.PersistentFlags().StringVarP(&config.ServerAddress, "host", "H", "localhost:8080", "address:port to listen on (defaults to localhost:8080)")
// 	// ServerCmd.PersistentFlags().StringVarP(&config.PathToDB, "db-file", "", "./exeguo.sqlite", "Path to the db file. Will be created if non-existant.")
// }

// func GenerateRsaKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
// 	privkey, _ := rsa.GenerateKey(rand.Reader, 4096)
// 	return privkey, &privkey.PublicKey
// }

// func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
// 	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
// 	privkey_pem := pem.EncodeToMemory(
// 		&pem.Block{
// 			Type:  "RSA PRIVATE KEY",
// 			Bytes: privkey_bytes,
// 		},
// 	)
// 	return string(privkey_pem)
// }

// func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
// 	block, _ := pem.Decode([]byte(privPEM))
// 	if block == nil {
// 		return nil, errors.New("failed to parse PEM block containing the key")
// 	}

// 	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return priv, nil
// }

// func ExportRsaPublicKeyAsPemStr(pubkey *rsa.PublicKey) (string, error) {
// 	pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)
// 	if err != nil {
// 		return "", err
// 	}
// 	pubkey_pem := pem.EncodeToMemory(
// 		&pem.Block{
// 			Type:  "RSA PUBLIC KEY",
// 			Bytes: pubkey_bytes,
// 		},
// 	)

// 	return string(pubkey_pem), nil
// }

// func ParseRsaPublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
// 	block, _ := pem.Decode([]byte(pubPEM))
// 	if block == nil {
// 		return nil, errors.New("failed to parse PEM block containing the key")
// 	}

// 	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	switch pub := pub.(type) {
// 	case *rsa.PublicKey:
// 		return pub, nil
// 	default:
// 		break // fall through
// 	}
// 	return nil, errors.New("Key type is not RSA")
// }

var GenCredentialsCmd = &cobra.Command{
	Use:   "auth",
	Short: "Generate auth data",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		serverAddress := "127.0.0.1"
		rootCertData, err := security.GenerateRootCertAndKey()

		if err != nil {
			log.Fatalf("generating root cert/key: %v", err)
		}

		serverCertData, err := security.GenerateServerCertAndKey(rootCertData, serverAddress)

		if err != nil {
			log.Fatalf("generating server cert/key: %v", err)
		}

		clientCertData, err := security.GenerateClientCertAndKey(rootCertData)

		if err != nil {
			log.Fatalf("generating server cert/key: %v", err)
		}

		fmt.Printf("CA key:\n%s\n", rootCertData.PrivateKeyPEM)
		fmt.Printf("CA cert:\n%s\n", rootCertData.CertPEM)

		fmt.Printf("Server key:\n%s\n", serverCertData.PrivateKeyPEM)
		fmt.Printf("Server cert:\n%s\n", serverCertData.CertPEM)

		fmt.Printf("Client key:\n%s\n", clientCertData.PrivateKeyPEM)
		fmt.Printf("Client cert:\n%s\n", clientCertData.CertPEM)
	},
}
