package account

import (
	"encoding/json"
	"errors"
	"github.com/hyperchain/gosdk/common"
	"github.com/hyperchain/gosdk/common/math"
	"github.com/hyperchain/gosdk/utils/ecdsa"
	"github.com/hyperchain/gosdk/utils/encrypt"
	"github.com/hyperchain/gosdk/utils/encrypt/guomi"
	"github.com/hyperchain/gosdk/utils/gm"
)

var logger = common.GetLogger("account")

type accountJSON struct {
	Address common.Address `json:"address"`
	// Algo 0x01 KDF2 0x02 DES(ECB) 0x03(plain) 0x04 DES(CBC)
	Algo                string `json:"algo,omitempty"`
	Encrypted           string `json:"encrypted,omitempty"`
	Version             string `json:"version,omitempty"`
	PublicKey           string `json:"publicKey,omitempty"`
	PrivateKey          string `json:"privateKey,omitempty"`
	PrivateKeyEncrypted bool   `json:"privateKeyEncrypted"`
}

// NewAccount create account using ecdsa
// if password is empty, the encrypted field will be private key
func NewAccount(password string) (string, error) {
	key, err := encrypt.GenerateKey()
	if err != nil {
		return "", err
	}
	accountJSON := new(accountJSON)

	// 私钥加密
	var encrypted []byte
	if password != "" {
		accountJSON.Algo = "0x02"
		accountJSON.Version = "1.0"
		encrypted, err = encrypt.DesEncrypt(math.PaddedBigBytes(key.D, 32), []byte(password))
		if err != nil {
			return "", err
		}
		accountJSON.PrivateKeyEncrypted = true
	} else {
		accountJSON.Algo = "0x03"
		accountJSON.Version = "2.0"
		encrypted = math.PaddedBigBytes(key.D, 32)
	}

	accountJSON.Address = encrypt.PubkeyToAddress(key.PublicKey)
	accountJSON.Encrypted = common.Bytes2Hex(encrypted)

	jsonBytes, err := json.Marshal(accountJSON)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// NewAccountFromPriv 从私钥字节数组得到ECDSA Key结构体
func NewAccountFromPriv(priv string) (*ecdsa.Key, error) {
	rawKey := encrypt.ToECDSA(common.Hex2Bytes(priv))
	if rawKey == nil {
		logger.Error("create account error")
		return nil, errors.New("create account error")
	}
	key := ecdsa.Key{
		PrivateKey: rawKey,
		PublicKey:  &rawKey.PublicKey,
	}

	return &key, nil
}

// NewAccountFromAccountJSON ECDSA Key结构体
func NewAccountFromAccountJSON(accountjson, password string) (key *ecdsa.Key, err error) {
	defer func() {
		if r := recover(); r != nil {
			key = nil
			err = errors.New("decrypt private key failed")
		}
	}()
	account := new(accountJSON)
	err = json.Unmarshal([]byte(accountjson), account)
	if err != nil {
		return nil, err
	}

	var priv []byte

	if account.Version == "1.0" {
		priv, err = encrypt.DesDecrypt(common.Hex2Bytes(account.Encrypted), []byte(password))
		if err != nil {
			return nil, err
		}
	} else {
		// version 2.0 means not encrypted
		priv = common.Hex2Bytes(account.Encrypted)
	}

	return NewAccountFromPriv(common.Bytes2Hex(priv))
}

// NewAccountSm2 生成国密
func NewAccountSm2(password string) (string, error) {
	key, err := gm.GenerateKey()
	if err != nil {
		return "", err
	}

	accountJson := new(accountJSON)
	temKey := key.Key
	tep := []byte{0}
	temKey = append(tep, temKey...)
	var privateKey []byte
	if password != "" {
		privateKey, err = encrypt.DesEncrypt(temKey, []byte(password))
		if err != nil {
			return "", err
		}
		accountJson.PrivateKeyEncrypted = true
	} else {
		privateKey = temKey
		accountJson.PrivateKeyEncrypted = false
	}
	accountJson.PrivateKey = common.Bytes2Hex(privateKey)
	pubKey := guomi.GetPubKeyFromPri(key)

	accountJson.PublicKey = common.Bytes2Hex(pubKey)
	accountJson.Address = common.BytesToAddress(gm.Keccak256(pubKey[0:])[12:])
	jsonBytes, err := json.Marshal(accountJson)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// NewAccountSm2FromPriv 从私钥字符串生成国密结构体
func NewAccountSm2FromPriv(priv string) (*gm.Key, error) {
	prk, err := guomi.GetPriKeyFromHex(common.Hex2Bytes(priv))
	if err != nil {
		return nil, err
	}
	key := gm.Key{
		PrivateKey: prk,
		PublicKey:  &prk.PublicKey,
	}

	return &key, nil
}

// NewAccountSm2FromAccountJSON 从账户JSON转为国密结构体
func NewAccountSm2FromAccountJSON(accountjson, password string) (key *gm.Key, err error) {
	defer func() {
		if r := recover(); r != nil {
			key = nil
			err = errors.New("decrypt private key failed")
		}
	}()
	account := new(accountJSON)
	err = json.Unmarshal([]byte(accountjson), account)
	if err != nil {
		return nil, err
	}
	var priv []byte
	if account.PrivateKeyEncrypted {
		priv, err = encrypt.DesDecrypt(common.Hex2Bytes(account.PrivateKey), []byte(password))
		if err != nil {
			return nil, err
		}
	} else {
		priv = common.Hex2Bytes(account.PrivateKey)
	}
	return NewAccountSm2FromPriv(common.Bytes2Hex(priv))
}
