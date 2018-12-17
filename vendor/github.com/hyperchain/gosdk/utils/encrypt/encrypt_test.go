package encrypt

import (
	"fmt"
	"github.com/hyperchain/gosdk/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrivateToAddress(t *testing.T) {
	a, _ := PrivateToAddress("0xcf66f23e76f08c452ef21872b592d0ef8f9331d158f8bca49ac36a0bdb052f21")
	if a != "0xeb28073ec2581727731805baab2fcbd13ea83b3f" {
		t.Error("address:", a)
	}
}

func TestEncryptKey(t *testing.T) {
	key := []byte("sfe023f_")
	result, err := DesEncrypt([]byte("tangkaikai@hyperchain.cn"), key)
	if err != nil {
		panic(err)
	}
	//fmt.Println(base64.StdEncoding.EncodeToString(result))
	original, err := DesDecrypt(result, key)
	if err != nil {
		panic(err)
	}
	assert.EqualValues(t, "tangkaikai@hyperchain.cn", string(original))
}

func TestDesEncrypt(t *testing.T) {
	data, _ := DesDecrypt([]byte("tomkk"), []byte("12345678"))
	fmt.Println(common.Bytes2Hex(data))
}
