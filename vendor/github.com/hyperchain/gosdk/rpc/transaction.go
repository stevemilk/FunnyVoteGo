package rpc

import (
	"encoding/hex"
	"errors"
	"github.com/hyperchain/gosdk/abi"
	"github.com/hyperchain/gosdk/common"
	"github.com/hyperchain/gosdk/utils/ecdsa"
	"github.com/hyperchain/gosdk/utils/encrypt"
	"github.com/hyperchain/gosdk/utils/gm"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	kec256Hash = encrypt.NewKeccak256Hash("keccak256")
	encryption = encrypt.NewEcdsaEncrypto("ecdsa")
)

// VMType vm type, could by evm and jvm
type VMType string

// VMType vm type, could by evm and jvm for now
const (
	EVM VMType = "EVM"
	JVM VMType = "JVM"
	//JSVM VMType = "jsvm"
)

// Params interface
type Params interface {
	// Serialize serialize to map
	Serialize() interface{}
	// SerializeToString serialize to string
	SerializeToString() string
}

// Transaction transaction entity
type Transaction struct {
	from       string
	to         string
	value      int64
	payload    string
	timestamp  int64
	nonce      int64
	signature  string
	opcode     int64
	vmType     string
	simulate   bool
	isValue    bool
	isDeploy   bool
	isMaintain bool
	isInvoke   bool
	extra      string
	hasExtra   bool
}

// NewTransaction return a empty transaction
func NewTransaction(from string) *Transaction {
	return &Transaction{
		timestamp: getCurTimeStamp(),
		nonce:     getRandNonce(),
		to:        "0x0",
		from:      chPrefix(from),
		simulate:  false,
		vmType:    string(EVM),
	}
}

// Simulate add transaction simulate
func (t *Transaction) Simulate(simulate bool) *Transaction {
	t.simulate = simulate
	return t
}

// VMType add transaction vmType
func (t *Transaction) VMType(vmType VMType) *Transaction {
	t.vmType = string(vmType)
	return t
}

// Transfer transfer balance to account
func (t *Transaction) Transfer(to string, value int64) *Transaction {
	t.value = value
	t.to = chPrefix(to)
	t.isValue = true
	return t
}

// Maintain maintain contract transaction
func (t *Transaction) Maintain(op int64, to, payload string) *Transaction {
	t.opcode = op
	t.payload = chPrefix(payload)
	t.to = chPrefix(to)
	t.isMaintain = true
	return t
}

// Deploy add transaction isDeploy
func (t *Transaction) Deploy(payload string) *Transaction {
	t.payload = chPrefix(payload)
	t.isDeploy = true
	return t
}

// DeployArgs add transaction deploy args
func (t *Transaction) DeployArgs(abiString string, args ...interface{}) *Transaction {
	ABI, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		logger.Error(err)
		return nil
	}

	packed, err := ABI.Pack("", args...)
	if err != nil {
		logger.Error(err)
		return nil
	}
	t.payload = t.payload + hex.EncodeToString(packed)
	t.isDeploy = true
	return t
}

// Invoke add transaction isInvoke
func (t *Transaction) Invoke(to string, payload []byte) *Transaction {
	t.payload = chPrefix(common.Bytes2Hex(payload))
	t.to = chPrefix(to)
	t.isInvoke = true
	return t
}

// Extra add extra into transaction
func (t *Transaction) Extra(extra string) *Transaction {
	t.extra = extra
	t.hasExtra = true
	return t
}

// To add transaction to
func (t *Transaction) To(to string) *Transaction {
	t.to = chPrefix(to)
	return t
}

// Payload add transaction payload
func (t *Transaction) Payload(payload string) *Transaction {
	t.payload = chPrefix(payload)
	return t
}

// Value add transaction value
func (t *Transaction) Value(value int64) *Transaction {
	t.value = value
	t.isValue = true
	return t
}

// OpCode add transaction opCode
func (t *Transaction) OpCode(op int64) *Transaction {
	t.opcode = op
	t.isMaintain = true
	return t
}

// needHashString construct a stirng that need to hash
func needHashString(t *Transaction) string {
	var (
		needHash string
		value    string
	)

	if t.isValue {
		value = "0x" + strconv.FormatInt(t.value, 16)
	} else if (t.isMaintain && t.opcode != 1) || t.payload == "" {
		value = "0x0"
	} else {
		value = strings.ToLower(common.StringToHex(t.payload))
	}

	needHash = "from=" + common.StringToHex(strings.ToLower(t.from)) +
		"&to=" + common.StringToHex(strings.ToLower(t.to)) +
		"&value=" + value +
		"&timestamp=0x" + strconv.FormatInt(t.timestamp, 16) +
		"&nonce=0x" + strconv.FormatInt(t.nonce, 16) +
		"&opcode=" + strconv.FormatInt(t.opcode, 16) +
		"&extra=" + t.extra +
		"&vmtype=" + t.vmType

	return needHash
}

// Sign support ecdsa or SM2 signature
//
// |    TYPE    |  ALGORITHM
// |------------|-------------
// | *ecdsa.Key |  ecdsa
// | *gm.Key    |  SM2
func (t *Transaction) Sign(key interface{}) {
	switch key.(type) {
	case *ecdsa.Key:
		ecdsaKey := key.(*ecdsa.Key)
		needHash := needHashString(t)
		sig, err := genSignature(nil, ecdsaKey.GetPrivKey(), "ecdsa", needHash)
		if err != nil {
			logger.Error("ecdsa signature error")
			return
		}
		t.signature = common.ToHex(sig)
		logger.Info(t)
	case *gm.Key:
		gmKey := key.(*gm.Key)
		needHash := needHashString(t)
		sig, err := genSignature(gmKey, "", "sm", needHash)
		if err != nil {
			logger.Error("sm2 signature error")
			return
		}
		t.signature = common.ToHex(sig)
	default:
		logger.Error("unsupported sign type")
	}
}

// genSignature get a signature of gm or ecdsa
func genSignature(gmKey *gm.Key, privateKey string, signatureType string, needHash string) ([]byte, StdError) {
	switch signatureType {
	case "ecdsa":
		logger.Info("sign type : ecdsa")
		sig, err := encrypt.Secp256k1Sign(hashForSign(needHash, kec256Hash).Bytes(), common.FromHex(privateKey))
		if err != nil {
			return nil, NewSystemError(err)
		}
		return sig, nil
	case "sm":
		logger.Info("sign type : sm")
		sig, err := gm.GenerateSignature(gmKey.GetPrivateKey(), gmKey.GetPublicKey(), gmKey.GetRawPublicKey(), needHash)
		if err != nil {
			return nil, NewSystemError(err)
		}
		return sig, nil
	default:
		return nil, NewSystemError(errors.New("signature type error"))
	}
}

// hashForSign get hash of a string for signing
func hashForSign(needHash string, ch encrypt.CommonHash) common.Hash {
	hashResult := ch.ByteHash([]byte(needHash))
	return hashResult
}

// getCurTimeStamp get current timestamp
func getCurTimeStamp() int64 {
	return time.Now().UnixNano()
}

// getRandNonce get a random nonce
func getRandNonce() int64 {
	return rand.Int63()
}

// chPrefix return a string start with '0x'
func chPrefix(origin string) string {
	if strings.HasPrefix(origin, "0x") {
		return origin
	}
	return "0x" + origin
}

// Serialize serialize the tx instance to a map
func (t *Transaction) Serialize() interface{} {
	if t.signature == "" {
		logger.Warning("this transaction is not signature")
	}
	param := make(map[string]interface{})
	param["from"] = t.from

	if !t.isDeploy || t.isMaintain {
		param["to"] = t.to
	}

	param["timestamp"] = t.timestamp
	param["nonce"] = t.nonce

	if !t.isMaintain {
		param["simulate"] = t.simulate
	}

	param["type"] = t.vmType

	if t.isValue {
		param["value"] = t.value
	} else if t.isMaintain && (t.opcode == 2 || t.opcode == 3) {

	} else {
		param["payload"] = t.payload
	}

	param["signature"] = t.signature

	if t.isMaintain {
		param["opcode"] = t.opcode
	}

	if t.hasExtra {
		param["extra"] = t.extra
	}

	return param
}

// SerializeToString serialize the tx instance to json string
func (t *Transaction) SerializeToString() string {
	return ""
}

// ChangeTxStruct change field value in transaction for phone scan sign
func (t *Transaction) ChangeTxStruct(change map[string]interface{}) {
	if change["from"] != "" {
		t.from = change["from"].(string)
	}
	if change["to"] != "" {
		t.to = change["to"].(string)
	}
	if change["value"] != 0 {
		t.value = change["value"].(int64)
	}
	if change["payload"] != "" {
		t.payload = change["payload"].(string)
	}
	if change["timestamp"] != 0 {
		t.timestamp = change["timestamp"].(int64)
	}
	if change["nonce"] != 0 {
		t.nonce = change["nonce"].(int64)
	}
	if change["signature"] != "" {
		t.signature = change["signature"].(string)
	}
	if change["opcode"] != 0 {
		t.opcode = change["opcode"].(int64)
	}
	if change["vmType"] != "" {
		t.vmType = change["vmType"].(string)
	}

}
