package security

import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
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

type AgentAuth struct {
	KeyPair *tls.Certificate
	CACert  *x509.Certificate
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

func GenerateAgentPEM(caCert *x509.Certificate, caKey *rsa.PrivateKey, address string) ([]byte, error) {
	keyPair, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	writer := bufio.NewWriter(&output)

	certTmpl, err := CertTemplate()

	if err != nil {
		return nil, err
	}

	certTmpl.KeyUsage = x509.KeyUsageDigitalSignature
	certTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}

	if ip := net.ParseIP(address); ip != nil {
		certTmpl.IPAddresses = append(certTmpl.IPAddresses, ip)
	} else {
		certTmpl.DNSNames = append(certTmpl.DNSNames, address)
	}

	certDER, err := x509.CreateCertificate(rand.Reader, certTmpl, caCert, &keyPair.PublicKey, caKey)

	if err != nil {
		return nil, err
	}

	pem.Encode(writer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	pem.Encode(writer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	pem.Encode(writer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(keyPair),
	})

	writer.Flush()

	return output.Bytes(), nil
}

func ParseAgentPEM(pemBytes []byte) (*AgentAuth, error) {
	var block *pem.Block
	block, pemBytes = pem.Decode(pemBytes)

	result := &AgentAuth{
		CACert:  &x509.Certificate{},
		KeyPair: &tls.Certificate{},
	}

	for ; block != nil; block, pemBytes = pem.Decode(pemBytes) {
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}

			if cert.IsCA {
				result.CACert = cert
			} else {
				result.KeyPair.Certificate = [][]byte{cert.Raw}
			}
		} else if block.Type == "RSA PRIVATE KEY" {
			rsaPrivKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("could not parse DER encoded key: %v", err)
			}
			result.KeyPair.PrivateKey = rsaPrivKey
		} else {
			return nil, fmt.Errorf("invalid pem block type: %s", block.Type)
		}
	}

	return result, nil
}

func LoadCertficateAndKeyFromFile(path string) (*tls.Certificate, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cert tls.Certificate
	for {
		block, rest := pem.Decode(raw)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, block.Bytes)
		} else {
			cert.PrivateKey, err = parsePrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("Failure reading private key from \"%s\": %s", path, err)
			}
		}
		raw = rest
	}

	if len(cert.Certificate) == 0 {
		return nil, fmt.Errorf("No certificate found in \"%s\"", path)
	} else if cert.PrivateKey == nil {
		return nil, fmt.Errorf("No private key found in \"%s\"", path)
	}

	return &cert, nil
}

func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("Found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("Failed to parse private key")
}
