package rpc

import (
	"crypto/ecdsa"
	"errors"
	"github.com/hyperchain/gosdk/common"
	"github.com/hyperchain/gosdk/utils/encrypt"
	"github.com/hyperchain/gosdk/utils/encrypt/guomi"
	"github.com/hyperchain/gosdk/utils/gm"
	"github.com/terasum/viper"
	"io/ioutil"
	"strings"
)

// KeyPair privateKey(ecdsa.PrivateKey or guomi.PrivateKey) and publicKey string
type KeyPair struct {
	privKey interface{}
	pubKey  string
}

// newKeyPair create a new KeyPair(ecdsa or sm2)
func newKeyPair(privFilePath string) (*KeyPair, error) {
	keyPari, err := encrypt.ParsePriv(privFilePath)
	if err != nil {
		logger.Debug("the cert is not ecdsa, now try to parse by sm2")
		keyPari, err := gm.GetPrivateKey(privFilePath)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		pubKey := guomi.GetPubKeyFromPri(keyPari)
		return &KeyPair{
			privKey: keyPari,
			pubKey:  common.Bytes2Hex(pubKey),
		}, nil
	}
	return &KeyPair{
		privKey: keyPari,
		pubKey:  common.Bytes2Hex(encrypt.FromECDSAPub(&keyPari.PublicKey)),
	}, nil
}

// Sign sign the message by privateKey
func (key *KeyPair) Sign(msg []byte) ([]byte, error) {
	switch key.privKey.(type) {
	case *ecdsa.PrivateKey:
		data, err := encrypt.ECDSASignWithSha256(key.privKey.(*ecdsa.PrivateKey), msg)
		if err != nil {
			return nil, err
		}
		return data, nil
	case *guomi.PrivateKey:
		gmKey := key.privKey.(*guomi.PrivateKey)
		data, err := gmKey.Sign(gm.SighHashSM3(gmKey.X, gmKey.Y, string(msg)))
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		logger.Error("unsupported sign type")
		return nil, NewSystemError(errors.New("signature type error"))
	}
}

// TCert tcert message
type TCert string

// TCertManager manager tcert
type TCertManager struct {
	sdkCert        *KeyPair
	uniqueCert     *KeyPair
	ecert          string
	tcertPool      map[string]TCert
	sdkcertPath    string
	sdkcertPriPath string
	uniquePubPath  string
	uniquePrivPath string
	cfca           bool
}

// NewTCertManager create a new TCert manager
func NewTCertManager(vip *viper.Viper, confRootPath string) (*TCertManager, error) {
	sdkcertPath := strings.Join([]string{confRootPath, vip.GetString(common.PrivacySDKcertPath)}, "/")
	logger.Debugf("[CONFIG]: sdkcertPath = %v", sdkcertPath)

	sdkcertPriPath := strings.Join([]string{confRootPath, vip.GetString(common.PrivacySDKcertPrivPath)}, "/")
	logger.Debugf("[CONFIG]: sdkcertPriPath = %v", sdkcertPriPath)

	uniquePubPath := strings.Join([]string{confRootPath, vip.GetString(common.PrivacyUniquePubPath)}, "/")
	logger.Debugf("[CONFIG]: uniquePubPath = %v", uniquePubPath)

	uniquePrivPath := strings.Join([]string{confRootPath, vip.GetString(common.PrivacyUniquePrivPath)}, "/")
	logger.Debugf("[CONFIG]: uniquePrivPath = %v", uniquePrivPath)

	cfca := vip.GetBool(common.Cfca)
	logger.Debugf("[CONFIG]: cfca = %v", cfca)

	var (
		sdkCert    *KeyPair
		uniqueCert *KeyPair
		err        error
	)

	sdkCert, err = newKeyPair(sdkcertPriPath)
	if err != nil {
		return nil, err
	}
	uniqueCert, err = newKeyPair(uniquePrivPath)
	if err != nil {
		return nil, err
	}
	ecert, err := ioutil.ReadFile(sdkcertPath)
	if err != nil {
		return nil, err
	}

	return &TCertManager{
		sdkcertPath:    sdkcertPath,
		sdkcertPriPath: sdkcertPriPath,
		uniquePubPath:  uniquePubPath,
		uniquePrivPath: uniquePrivPath,
		sdkCert:        sdkCert,
		uniqueCert:     uniqueCert,
		ecert:          common.Bytes2Hex(ecert),
	}, nil
}

// GetECert get ecert
func (tcm *TCertManager) GetECert() string {
	return tcm.ecert
}
