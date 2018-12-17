package cert

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"hyperbaas/src/lib/certgen/primitives"
	"io/ioutil"
	"os"
)

// SelfSignCA 自己生成根证书 ok
func SelfSignCA(selfCAPath string, selfCAPrivPath string) error {
	der, pri, err := primitives.NewSelfSignedCert()
	if err != nil {
		return err
	}
	certPemByte := primitives.DERCertToPEM(der)
	file, err := os.Create(selfCAPath)
	if err != nil {
		return err
	}
	file.WriteString(string(certPemByte))

	var block pem.Block
	block.Type = "EC PRIVATE KEY"
	priv := pri.(*ecdsa.PrivateKey)
	der, err = primitives.PrivateKeyToDER(priv)
	if err != nil {
		return err
	}

	block.Bytes = der
	file, err = os.Create(selfCAPrivPath)
	if err != nil {
		return err
	}
	pem.Encode(file, &block)
	return nil
}

// CheckCertSignature 检查pem格式的证书的合法性 ok
func CheckCertSignature(certPath string) error {
	fileContent, err := ioutil.ReadFile(certPath)
	if err != nil {
		return err
	}
	certStr := string(fileContent)
	block, _ := pem.Decode([]byte(certStr))
	cert, _ := x509.ParseCertificate(block.Bytes)
	err = cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature)
	return err
}

// GeneratePrivKeyFile 生成私钥pem格式文件 ok
func GeneratePrivKeyFile(privPath string) error {
	pri, _ := primitives.NewECDSAKey()
	fmt.Println("===============")
	fmt.Println("生成私钥为：")
	fmt.Println(pri)
	fmt.Println("===============")

	//fmt.Println(json)
	var block pem.Block
	block.Type = "EC PRIVATE KEY"
	der, err := primitives.PrivateKeyToDER(pri)
	if err != nil {
		return err
	}
	block.Bytes = der
	file, err := os.Create(privPath)
	if err != nil {
		return err
	}
	pem.Encode(file, &block)
	return nil
}

// GeneratePrivKey 生成私钥pem格式内容
func GeneratePrivKey() ([]byte, error) {
	pri, _ := primitives.NewECDSAKey()
	fmt.Println("===============")
	fmt.Println("生成私钥为：")
	fmt.Println(pri)
	fmt.Println("===============")

	//fmt.Println(json)
	var block pem.Block
	block.Type = "EC PRIVATE KEY"
	der, err := primitives.PrivateKeyToDER(pri)
	if err != nil {
		return nil, err
	}
	block.Bytes = der
	pribyte := pem.EncodeToMemory(&block)
	return pribyte, nil
}

// ParsePrivateKey 解析PEM格式的私钥
func ParsePrivateKey(privPath string) (key interface{}, err error) {
	content, _ := ioutil.ReadFile(privPath)
	privateKey := string(content)
	block, _ := pem.Decode([]byte(privateKey))
	//var pri ecdsa.PrivateKey
	return primitives.DERToPrivateKey(block.Bytes)
}

// CreateCert 生成证书 ok
func CreateCert(rootCertPath string, rootCertPrivPath string, targetCertPath string, tarCertPrivPath string, isCA bool) error {
	rootCertByte, err := ioutil.ReadFile(rootCertPath)
	if err != nil {
		return err
	}

	rootCert, err := primitives.ParseCertificate(string(rootCertByte))
	if err != nil {
		return err
	}
	rootCertPrivateKeyByte, err := ioutil.ReadFile(rootCertPrivPath)
	if err != nil {
		return err
	}

	rootCertPrivateKeyBlock, _ := pem.Decode(rootCertPrivateKeyByte)

	rootCertPrivateKey, err := primitives.DERToPrivateKey(rootCertPrivateKeyBlock.Bytes)
	if err != nil {
		return err
	}

	certByCa, _, privateKeyPemBlockByCa, err := primitives.CreateCertByCa(rootCert, rootCertPrivateKey, isCA)
	if err != nil {
		return err
	}

	certByCaPem := primitives.DERCertToPEM(certByCa)
	//写cert文件
	file, err := os.Create(targetCertPath)
	if err != nil {
		return err
	}
	file.WriteString(string(certByCaPem))
	//写私钥文件
	privfile, err := os.Create(tarCertPrivPath)
	if err != nil {
		return err
	}
	privfile.WriteString(string(privateKeyPemBlockByCa))

	//验证证书
	certByCaX509, _ := primitives.DERToX509Certificate(certByCa)
	//ecertByCa1 := ParseCertificate()
	//pub1 := certByCaX509.PublicKey.(*ecdsa.PublicKey)
	//fmt.Println(*pub1)
	fmt.Println("---------PAST DUE---------")
	fmt.Println("subcert: ", certByCaX509.NotAfter)
	fmt.Println("rootcert: ", rootCert.NotAfter)
	fmt.Println("--------------------------")
	err = certByCaX509.CheckSignatureFrom(rootCert)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// SignPayload 签名
func SignPayload(certPrivPath string, payload []byte) ([]byte, error) {

	ee := primitives.NewEcdsaEncrypto("ecdsa")
	//payload := []byte{1,2,3}

	certPrivContent, err := ioutil.ReadFile(certPrivPath)
	if err != nil {
		return nil, err
	}

	pri, err := primitives.ParsePriKey(string(certPrivContent))
	if err != nil {
		return nil, err
	}

	//fmt.Println(pri)
	sign, err := ee.Sign(payload, pri)

	if err != nil {
		return nil, err
	}
	return sign, nil
}

// VerifyPayload 测试签名
func VerifyPayload(certPath string, signedPayload []byte, originPayload []byte) (bool, error) {

	certContent, err := ioutil.ReadFile(certPath)
	if err != nil {
		return false, err
	}
	cert, err := primitives.ParseCertificate(string(certContent))
	if err != nil {
		return false, err
	}
	pub := cert.PublicKey

	return primitives.ECDSAVerify(pub, originPayload, signedPayload)
}

// CheckCert 检查证书
func CheckCert(subCertPath, parentCertPath string) (bool, error) {
	subCertContent, err := ioutil.ReadFile(subCertPath)
	if err != nil {
		return false, err
	}
	parentCertContent, err := ioutil.ReadFile(parentCertPath)
	if err != nil {
		return false, err
	}

	perentCert, err := primitives.ParseCertificate(string(parentCertContent))
	if err != nil {
		return false, err
	}
	subCert, err := primitives.ParseCertificate(string(subCertContent))
	if err != nil {
		return false, err
	}
	err = subCert.CheckSignatureFrom(perentCert)
	if err != nil {
		return false, err
	}
	return true, nil
}
