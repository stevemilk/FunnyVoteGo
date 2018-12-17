package account

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAccount(t *testing.T) {
	accountjson := `{"address":"0x534fca8ee67ce07a45658e02b37a3164d8004cc5","algo":"0x01","encrypted":"ed73b138f4ec72ac85110514125d3ab9edac57e7bd8e19363697925cecea768bfeb959b7d4642fcb","version":"1.0"}`
	key, err := NewAccountFromAccountJSON(accountjson, "1234567890")
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, "f48dd35986fdb804556b299f08aa61153c68ea378492df3dd848513677e4b3c2", key.GetPrivKey(), "私钥解析错误")
}

func TestNewAccountFromPriv(t *testing.T) {
	privateKey, err := NewAccountFromPriv("a1fd6ed6225e76aac3884b5420c8cdbb4fde1db01e9ef773415b8f2b5a9b77d4")
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, "a1fd6ed6225e76aac3884b5420c8cdbb4fde1db01e9ef773415b8f2b5a9b77d4", privateKey.GetPrivKey(), "私钥解析错误")
}

func TestNewAccountECDSA(t *testing.T) {
	accountJson, err := NewAccount("123")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = NewAccountFromAccountJSON(accountJson, "123")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(accountJson)
}

func TestNewAccountSm2(t *testing.T) {
	account, err := NewAccountSm2("123")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = NewAccountSm2FromAccountJSON(account, "123")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(account)
}

func TestNewAccountSm2FromPriv(t *testing.T) {
	key, _ := NewAccountSm2FromPriv("b15a43adb0bccef47fbe8d716a0b5c616c54f879242b101281ba82ab07ab0ddb")
	assert.EqualValues(t, "0x136e36a9996da1794c7582cdeba4f4852c218f78", key.GetAddress())
}

func TestNewAccountSm2FromAccountJSON(t *testing.T) {
	accountJson := `{"address":"0x8485147cbf02dec93ee84f81824a3b60e355f5cd","publicKey":"04a1b4c82a2a13e15a11e3ee9316504de0c3b54d46f5c189ae42603c9cd07a50fdca2ac35d0ceef4a8466ccb182f52403d9a58b573e1bf6fd4f52c31493bf7241b","privateKey":"f67136bf3caa4197a1cfaf38a5392ff94dae91bda700f8898b11cf49891a47bb","privateKeyEncrypted":false}`
	key, err := NewAccountSm2FromAccountJSON(accountJson, "")
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, "0x8485147cbf02dec93ee84f81824a3b60e355f5cd", key.GetAddress())
}

func TestNewAccountSm2FromAccountJSON2(t *testing.T) {
	accountJson, _ := NewAccountSm2("123")
	fmt.Println(accountJson)
	key, err := NewAccountSm2FromAccountJSON(accountJson, "123")
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, 42, len(key.GetAddress()))
}

func TestNewAccountSm2FromAccountJSON3(t *testing.T) {
	accountJson := `{"address":"0x503182ac93cbf5f1800856b81e8d2e8e773a757c","publicKey":"049a8e2b2b3089deca7b7081e695fdb31b7139eb64b34c20417bb0c8308c6134d74295f073ee2c0b541f974472597aba108f338d48a0f3f215b7075f9a31b55a9c","privateKey":"1749a9d89ae7304f88d517cb07b3c72e697bd38c627dda182b507e8474da2791df0d7bc08a13ae42","privateKeyEncrypted":true}`
	_, err := NewAccountSm2FromAccountJSON(accountJson, "321")

	assert.NotNil(t, err)
}

func Test(t *testing.T) {
	accountJSON, err := NewAccount("123")
	fmt.Println(accountJSON)
	if err != nil {
		t.Error(err)
	}
	_, err = NewAccountFromAccountJSON(accountJSON, "123")
	if err != nil {
		t.Error(err)
	}

}
