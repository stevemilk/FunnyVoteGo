package primitives

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"
)

var (
	defaultCurve elliptic.Curve
)

func init() {
	defaultCurve = elliptic.P256()
}

// GetDefaultCurve returns the default elliptic curve used by the crypto layer
func GetDefaultCurve() elliptic.Curve {
	return defaultCurve
}

// PrivateKeyToDER marshals a private key to der
func PrivateKeyToDER(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, ErrNilArgument
	}
	return x509.MarshalECPrivateKey(privateKey)
}

// DERToPrivateKey unmarshals a der to private key
func DERToPrivateKey(der []byte) (key interface{}, err error) {
	//fmt.Printf("DER [%s]\n", EncodeBase64(der))

	if key, err = x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	//fmt.Printf("DERToPrivateKey Err [%s]\n", err)
	if key, err = x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return
		default:
			return nil, errors.New("Found unknown private key type in PKCS#8 wrapping")
		}
	}
	//fmt.Printf("DERToPrivateKey Err [%s]\n", err)
	if key, err = x509.ParseECPrivateKey(der); err == nil {
		return
	}
	//fmt.Printf("DERToPrivateKey Err [%s]\n", err)

	return nil, errors.New("Failed to parse private key")
}

// DERCertToPEM converts der to pem
func DERCertToPEM(der []byte) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: der,
		},
	)
}

// DERToX509Certificate converts der to x509
func DERToX509Certificate(asn1Data []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(asn1Data)
}

// ParseCertificate 解析证书
func ParseCertificate(cert string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(cert))

	if block == nil {
		fmt.Println("failed to parse certificate PEM")
		return nil, errors.New("failed to parse certificate PEM")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		fmt.Println("faile to parse certificate")
		return nil, errors.New("faile to parse certificate")
	}

	return x509Cert, nil
}

// DERToPublicKey unmarshals a der to public key
func DERToPublicKey(derBytes []byte) (pub interface{}, err error) {
	key, err := x509.ParsePKIXPublicKey(derBytes)

	return key, err
}

// CreateCertByCa 创建通过ca证书签发新证书
func CreateCertByCa(ca *x509.Certificate, private interface{}, isCA bool) ([]byte, interface{}, []byte, error) {

	caPri := private.(*ecdsa.PrivateKey)

	newPrivKey, err := NewECDSAKey()

	//储存privateKey
	var block pem.Block
	block.Type = "EC PRIVATE KEY"
	der, _ := PrivateKeyToDER(newPrivKey)
	block.Bytes = der
	//file,_ := os.Create("tcert.priv")
	newPrivKeyPemByte := pem.EncodeToMemory(&block)

	if err != nil {
		return nil, nil, nil, err
	}

	extKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	unknownExtKeyUsage := []asn1.ObjectIdentifier{[]int{1, 2, 3}, []int{2, 59, 1}}
	extraExtensionData := []byte("extra extension")
	commonName := "hyperchain.cn"
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Hyperchain"},
			Country:      []string{"CHN"},
			ExtraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  []int{2, 5, 4, 42},
					Value: "Develop",
				},
				// This should override the Country, above.
				{
					Type:  []int{2, 5, 4, 6},
					Value: "ZH",
				},
			},
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(876000 * time.Hour), //暂定证书有效期为100年

		SignatureAlgorithm: x509.ECDSAWithSHA256,

		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageCertSign,

		ExtKeyUsage:        extKeyUsage,
		UnknownExtKeyUsage: unknownExtKeyUsage,

		BasicConstraintsValid: true,
		IsCA:                  isCA,

		ExtraExtensions: []pkix.Extension{
			{
				Id:    []int{1, 2, 3, 4},
				Value: extraExtensionData,
			},
		},
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, ca, &newPrivKey.PublicKey, caPri)
	if err != nil {
		return nil, nil, nil, err
	}

	return cert, newPrivKey, newPrivKeyPemByte, nil
}

// NewSelfSignedCert 生成自签名证书
func NewSelfSignedCert() ([]byte, interface{}, error) {
	privKey, err := NewECDSAKey()

	if err != nil {
		return nil, nil, err
	}

	testExtKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	testUnknownExtKeyUsage := []asn1.ObjectIdentifier{[]int{1, 2, 3}, []int{2, 59, 1}}
	extraExtensionData := []byte("extra extension")
	commonName := "hyperchain.cn"
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Hyperchain"},
			Country:      []string{"CHN"},
			ExtraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  []int{2, 5, 4, 42},
					Value: "Develop",
				},
				// This should override the Country, above.
				{
					Type:  []int{2, 5, 4, 6},
					Value: "ZH",
				},
			},
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(876000 * time.Hour), //暂定证书有效期为100年

		SignatureAlgorithm: x509.ECDSAWithSHA384,

		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageCertSign,

		ExtKeyUsage:        testExtKeyUsage,
		UnknownExtKeyUsage: testUnknownExtKeyUsage,

		BasicConstraintsValid: true,
		IsCA:                  true,

		//OCSPServer:            []string{"http://ocsp.example.com"},
		//IssuingCertificateURL: []string{"http://crt.example.com/ca1.crt"},

		//DNSNames:       []string{"test.example.com"},
		//EmailAddresses: []string{"gopher@golang.org"},
		//IPAddresses:    []net.IP{net.IPv4(127, 0, 0, 1).To4(), net.ParseIP("2001:4860:0:2001::68")},

		//PolicyIdentifiers:   []asn1.ObjectIdentifier{[]int{1, 2, 3}},
		//PermittedDNSDomains: []string{".example.com", "example.com"},

		//CRLDistributionPoints: []string{"http://crl1.example.com/ca1.crl", "http://crl2.example.com/ca1.crl"},

		ExtraExtensions: []pkix.Extension{
			{
				Id:    []int{1, 2, 3, 4},
				Value: extraExtensionData,
			},
		},
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, nil, err
	}

	return cert, privKey, nil
}

// ParsePriKey 解析PEM私钥
func ParsePriKey(derPri string) (interface{}, error) {
	block, _ := pem.Decode([]byte(derPri))

	pri, err1 := DERToPrivateKey(block.Bytes)

	if err1 != nil {
		return nil, err1
	}

	return pri, nil
}

// ParsePubKey 解析PEM公钥
func ParsePubKey(pubstr string) (interface{}, error) {
	//todo finish the public key parse

	block, _ := pem.Decode([]byte(pubstr))

	pub, err := DERToPublicKey(block.Bytes)

	if err != nil {
		return nil, err
	}
	return pub, nil

}
