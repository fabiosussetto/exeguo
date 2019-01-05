package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"log"
	"math/big"
	"net"
	"time"
)

type CertData struct {
	Cert          *x509.Certificate
	CertPEM       []byte
	PrivateKey    *rsa.PrivateKey
	PrivateKeyPEM []byte
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
		NotAfter:              time.Now().AddDate(5, 0, 0),
		BasicConstraintsValid: true,
	}
	return &tmpl, nil
}

func CreateCert(template, parent *x509.Certificate, pub interface{}, parentPriv interface{}) (
	cert *x509.Certificate, certPEM []byte, err error) {

	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, parentPriv)
	if err != nil {
		return nil, nil, err
	}
	// parse the resulting certificate so we can use it again
	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, err
	}
	// PEM encode the certificate (this is a standard TLS encoding)
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM = pem.EncodeToMemory(&b)
	return cert, certPEM, nil
}

func CreatePrivateKey() (*rsa.PrivateKey, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		log.Fatalf("generating random key: %v", err)
		return nil, nil, err
	}

	rootKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return key, rootKeyPEM, nil
}

func GenerateRootCertAndKey() (*CertData, error) {
	rootKey, rootKeyPEM, err := CreatePrivateKey()

	if err != nil {
		log.Fatalf("generating root key: %v", err)
		return nil, err
	}

	rootCertTmpl, err := CertTemplate()

	if err != nil {
		log.Fatalf("creating cert template: %v", err)
		return nil, err
	}
	// describe what the certificate will be used for
	rootCertTmpl.IsCA = true
	rootCertTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	rootCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	// rootCertTmpl.IPAddresses = []net.IP{net.ParseIP(ipAddress)}

	rootCert, rootCertPEM, err := CreateCert(rootCertTmpl, rootCertTmpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		log.Fatalf("error creating cert: %v", err)
		return nil, err
	}

	return &CertData{Cert: rootCert, CertPEM: rootCertPEM, PrivateKey: rootKey, PrivateKeyPEM: rootKeyPEM}, nil
}

func GenerateServerCertAndKey(caCertData *CertData, address string) (*CertData, error) {
	key, keyPEM, err := CreatePrivateKey()

	servCertTmpl, err := CertTemplate()
	if err != nil {
		return nil, err
	}

	servCertTmpl.KeyUsage = x509.KeyUsageDigitalSignature
	servCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}

	if ip := net.ParseIP(address); ip != nil {
		servCertTmpl.IPAddresses = append(servCertTmpl.IPAddresses, ip)
	} else {
		servCertTmpl.DNSNames = append(servCertTmpl.DNSNames, address)
	}

	// create a certificate which wraps the server's public key, sign it with the root private key
	cert, certPEM, err := CreateCert(servCertTmpl, caCertData.Cert, &key.PublicKey, caCertData.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &CertData{Cert: cert, CertPEM: certPEM, PrivateKey: key, PrivateKeyPEM: keyPEM}, nil
}

func GenerateClientCertAndKey(caCertData *CertData) (*CertData, error) {
	key, keyPEM, err := CreatePrivateKey()

	if err != nil {
		return nil, err
	}

	// create a template for the client
	clientCertTmpl, err := CertTemplate()

	if err != nil {
		return nil, err
	}

	clientCertTmpl.KeyUsage = x509.KeyUsageDigitalSignature
	clientCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}

	cert, certPEM, err := CreateCert(clientCertTmpl, caCertData.Cert, &key.PublicKey, caCertData.PrivateKey)

	if err != nil {
		return nil, err
	}

	return &CertData{Cert: cert, CertPEM: certPEM, PrivateKey: key, PrivateKeyPEM: keyPEM}, nil
}

func ParsePrivateKeyFromPEM(keyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(keyPEM)

	if block == nil || block.Type != "RSA PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing private key")
	}

	keyDER, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	return keyDER, nil
}

func ParseCertFromPEM(keyPEM []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(keyPEM)

	if block == nil || block.Type != "CERTIFICATE" {
		log.Fatal("failed to decode PEM block containing certificate")
	}

	certDER, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		return nil, err
	}

	return certDER, nil
}
