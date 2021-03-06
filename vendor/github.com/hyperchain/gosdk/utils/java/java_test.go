package java

import (
	"fmt"
	"github.com/coreos/etcd/pkg/testutil"
	"github.com/hyperchain/gosdk/common"
	"github.com/hyperchain/gosdk/rpc"
	"github.com/hyperchain/gosdk/utils/gm"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	hrpc = rpc.NewRPCWithPath("../../conf")

	guomiPub      = "02739518af5e065b22dabb35ea5369a4c64d4865565874a006399bbb0e62e18004"
	guomiPri      = "6153af264daa4763490f2a51c9d13417ef9f579229be2141574eb339ee9b9d2a"
	guomiKey, err = gm.GetKeyPareFromHex(guomiPri, guomiPub)

	contractAddress = "0x31cf62472b1856d94553d2fe78f3bb067afb0714"
)

type TestAsyncHandler struct {
	t        *testing.T
	IsCalled bool
}

func (tah *TestAsyncHandler) OnSuccess(receipt *rpc.TxReceipt) {
	tah.IsCalled = true
	fmt.Println(receipt.Ret)
}

func (tah *TestAsyncHandler) OnFailure(err rpc.StdError) {
	tah.t.Error(err.String())
}

func TestEncodeJavaFunc(t *testing.T) {
	res := EncodeJavaFunc("add", "tomkk", "tomkk")
	testutil.AssertEqual(t, "1206696e766f6b651a036164641a05746f6d6b6b1a05746f6d6b6b", common.Bytes2Hex(res))
}

func TestDecodeJavaResult(t *testing.T) {
	str := "Mr.汤"
	testutil.AssertEqual(t, "Mr.汤", DecodeJavaResult(common.Bytes2Hex([]byte(str))))
}

func TestDeployJavaContract(t *testing.T) {
	payload, err := ReadJavaContract("../../conf/contract/contract01")
	if err != nil {
		t.Error(err)
		return
	}

	tx := rpc.NewTransaction(guomiKey.GetAddress()).Deploy(payload).VMType(rpc.JVM)
	tx.Sign(guomiKey)
	asyncHandler := TestAsyncHandler{t: t}
	hrpc.DeployContractAsync(tx, &asyncHandler)
	time.Sleep(3 * time.Second)
	assert.EqualValues(t, true, asyncHandler.IsCalled, "回调未被执行")
}

func TestInvokeJavaContract(t *testing.T) {
	payload, err := ReadJavaContract("../../conf/contract/contract01")
	if err != nil {
		t.Error(err)
		return
	}
	tx := rpc.NewTransaction(guomiKey.GetAddress()).Deploy(payload).VMType(rpc.JVM)
	tx.Sign(guomiKey)
	txReceipt, stdErr := hrpc.DeployContract(tx)
	if stdErr != nil {
		t.Error(stdErr.String())
		return
	}

	contractAddress = txReceipt.ContractAddress

	tx = rpc.NewTransaction(guomiKey.GetAddress()).Invoke(contractAddress, EncodeJavaFunc("issue", guomiKey.GetAddress(), "1000")).VMType(rpc.JVM)

	tx.Sign(guomiKey)

	txReceipt, stdErr = hrpc.InvokeContract(tx)
	if stdErr != nil {
		t.Error(stdErr.String())
		return
	}

	fmt.Println(txReceipt.Ret)

	tx = rpc.NewTransaction(guomiKey.GetAddress()).Invoke(contractAddress, EncodeJavaFunc("getAccountBalance", guomiKey.GetAddress())).VMType(rpc.JVM)
	tx.Sign(guomiKey)

	txReceipt, stdErr = hrpc.InvokeContract(tx)
	if stdErr != nil {
		t.Error(stdErr.String())
		return
	}

	testutil.AssertEqual(t, "1000.0", DecodeJavaResult(txReceipt.Ret))
}

func TestDecodeJavaLog(t *testing.T) {
	payload, err := ReadJavaContract("../../conf/contract/contract01")
	if err != nil {
		t.Error(err)
		return
	}
	tx := rpc.NewTransaction(guomiKey.GetAddress()).Deploy(payload).VMType(rpc.JVM)
	tx.Sign(guomiKey)
	txReceipt, stdErr := hrpc.DeployContract(tx)
	if stdErr != nil {
		t.Error(stdErr.String())
		return
	}

	contractAddress = txReceipt.ContractAddress

	tx = rpc.NewTransaction(guomiKey.GetAddress()).Invoke(contractAddress, EncodeJavaFunc("testPostEvent", "TomKK")).VMType(rpc.JVM)

	tx.Sign(guomiKey)

	txReceipt, stdErr = hrpc.InvokeContract(tx)
	if stdErr != nil {
		t.Error(stdErr.String())
		return
	}
	res, err := DecodeJavaLog(txReceipt.Log[0].Data)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t,
		`{"name":"event0","atrributes":{"attr2":"value2","attr1":"value1","attr3":"value3"},"topics":["test","simulate_bank"]}`,
		res,
		"解码失败")
}

func Test(t *testing.T) {
	fmt.Println(DecodeJavaLog("65794a755957316c496a6f695a585a6c626e51774969776959585279636d6c696458526c6379493665794a68644852794d694936496e5a686248566c4d694973496d463064484978496a6f69646d4673645755784969776959585230636a4d694f694a32595778315a544d69665377696447397761574e7a496a7062496e526c633351694c434a7a61573131624746305a56396959573572496c3139"))
}
