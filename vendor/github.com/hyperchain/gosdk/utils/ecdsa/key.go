package ecdsa

import (
	"crypto/ecdsa"
	"github.com/hyperchain/gosdk/common"
	"github.com/hyperchain/gosdk/common/math"
	"github.com/hyperchain/gosdk/utils/encrypt"
)

type Key struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

func (key *Key) GetAddress() string {
	ret := encrypt.PubkeyToAddress(key.PrivateKey.PublicKey)
	return ret.Hex()
}

func (key *Key) GetPrivKey() string {
	return common.Bytes2Hex(math.PaddedBigBytes(key.PrivateKey.D, 32))
}
