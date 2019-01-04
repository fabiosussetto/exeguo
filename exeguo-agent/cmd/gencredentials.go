package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	// ServerCmd.PersistentFlags().StringVarP(&config.ServerAddress, "host", "H", "localhost:8080", "address:port to listen on (defaults to localhost:8080)")
	// ServerCmd.PersistentFlags().StringVarP(&config.PathToDB, "db-file", "", "./exeguo.sqlite", "Path to the db file. Will be created if non-existant.")
}

func GenerateRsaKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, _ := rsa.GenerateKey(rand.Reader, 4096)
	return privkey, &privkey.PublicKey
}

func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return string(privkey_pem)
}

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func ExportRsaPublicKeyAsPemStr(pubkey *rsa.PublicKey) (string, error) {
	pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return "", err
	}
	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return string(pubkey_pem), nil
}

func ParseRsaPublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break // fall through
	}
	return nil, errors.New("Key type is not RSA")
}

func CertTemplate() (*x509.Certificate, error) {
	// generate a random serial number (a real cert authority would have some logic behind this)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.New("failed to generate serial number: " + err.Error())
	}

	tmpl := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"Exeguo"}},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour), // valid for an hour
		BasicConstraintsValid: true,
	}
	return &tmpl, nil
}

func CreateCert(template, parent *x509.Certificate, pub interface{}, parentPriv interface{}) (
	cert *x509.Certificate, certPEM []byte, err error) {

	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, parentPriv)
	if err != nil {
		return
	}
	// parse the resulting certificate so we can use it again
	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return
	}
	// PEM encode the certificate (this is a standard TLS encoding)
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM = pem.EncodeToMemory(&b)
	return
}

var GenCredentialsCmd = &cobra.Command{
	Use:   "auth",
	Short: "Generate auth data",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// root CA cert
		// generate a new key-pair
		rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("generating random key: %v", err)
		}

		rootKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rootKey),
		})
		fmt.Printf("Root key:\n%s\n", rootKeyPEM)

		rootCertTmpl, err := CertTemplate()
		if err != nil {
			log.Fatalf("creating cert template: %v", err)
		}
		// describe what the certificate will be used for
		rootCertTmpl.IsCA = true
		rootCertTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
		rootCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
		rootCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

		rootCert, rootCertPEM, err := CreateCert(rootCertTmpl, rootCertTmpl, &rootKey.PublicKey, rootKey)
		if err != nil {
			log.Fatalf("error creating cert: %v", err)
		}

		fmt.Printf("Root cert:\n%s\n", rootCertPEM)
		// fmt.Printf("%#x\n", rootCert.Signature) // more ugly binary

		// create a key-pair for the server
		servKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("generating random key: %v", err)
		}

		// create a template for the server
		servCertTmpl, err := CertTemplate()
		if err != nil {
			log.Fatalf("creating cert template: %v", err)
		}
		servCertTmpl.KeyUsage = x509.KeyUsageDigitalSignature
		servCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		servCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

		// create a certificate which wraps the server's public key, sign it with the root private key
		_, servCertPEM, err := CreateCert(servCertTmpl, rootCert, &servKey.PublicKey, rootKey)
		if err != nil {
			log.Fatalf("error creating cert: %v", err)
		}

		// provide the private key and the cert
		servKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(servKey),
		})

		// servTLSCert, err := tls.X509KeyPair(servCertPEM, servKeyPEM)
		// if err != nil {
		// 	log.Fatalf("invalid key pair: %v", err)
		// }

		fmt.Printf("Server key:\n%s\n", servKeyPEM)
		fmt.Printf("Server cert:\n%s\n", servCertPEM)
		// create another test server and use the certificat

		// create a key-pair for the client
		clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("generating random key: %v", err)
		}

		// create a template for the client
		clientCertTmpl, err := CertTemplate()
		if err != nil {
			log.Fatalf("creating cert template: %v", err)
		}
		clientCertTmpl.KeyUsage = x509.KeyUsageDigitalSignature
		clientCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}

		// the root cert signs the cert by again providing its private key
		_, clientCertPEM, err := CreateCert(clientCertTmpl, rootCert, &clientKey.PublicKey, rootKey)
		if err != nil {
			log.Fatalf("error creating cert: %v", err)
		}

		// encode and load the cert and private key for the client
		clientKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey),
		})

		fmt.Printf("Client key:\n%s\n", clientKeyPEM)
		fmt.Printf("Client cert:\n%s\n", clientCertPEM)

		// clientTLSCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
		// if err != nil {
		// 	log.Fatalf("invalid key pair: %v", err)
		// }

		// fmt.Printf("%s\n", clientTLSCert)

	},
}
