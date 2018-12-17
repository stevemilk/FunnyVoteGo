package rpc

import (
	"fmt"
	"github.com/hyperchain/gosdk/abi"
	"strings"
	"testing"
	"time"
)

var (
	wsRPC = rpc
)

type TestEventHandler struct {
}

func (h *TestEventHandler) OnSubscribe() {
	fmt.Println("订阅成功！")
}

func (h *TestEventHandler) OnUnSubscribe() {
	fmt.Println("取消订阅成功！")
}

func (h *TestEventHandler) OnMessage(message []byte) {
	fmt.Printf("收到信息: %s\n", message)
}

func (h *TestEventHandler) OnClose() {
	fmt.Println("连接关闭回调调用！")
}

func TestWebSocketClient_BlockEvent(t *testing.T) {
	bf := NewBlockEventFilter()
	bf.BlockInfo = true
	wsCli := wsRPC.GetWebSocketClient()
	subID, err := wsCli.Subscribe(1, bf, &TestEventHandler{})
	if err != nil {
		t.Error(err.String())
		return
	}

	deployContract(binContract, address)

	time.Sleep(1 * time.Second)
	wsCli.UnSubscribe(subID)
	time.Sleep(1 * time.Second)
	wsCli.CloseConn(1)
	time.Sleep(1 * time.Second)
}

func TestWebSocketClient_SystemStatusEvent(t *testing.T) {
	sysf := NewSystemStatusFilter().
		AddModules("p2p").
		AddSubtypes("viewchange")
	wsCli := wsRPC.GetWebSocketClient()
	_, err := wsCli.Subscribe(1, sysf, &TestEventHandler{})
	if err != nil {
		t.Error(err.String())
		return
	}
	time.Sleep(1 * time.Second)
	wsCli.CloseConn(1)
	time.Sleep(1 * time.Second)
}

func TestWebSocketClient_LogsEvent(t *testing.T) {
	cr, _ := compileContract("../conf/contract/Accumulator2.sol")
	var arg [32]byte
	copy(arg[:], "test")
	ABI, _ := abi.JSON(strings.NewReader(cr.Abi[0]))
	transaction := NewTransaction(guomiKey.GetAddress()).Deploy(cr.Bin[0]).DeployArgs(cr.Abi[0], uint32(10), arg)
	transaction.Sign(guomiKey)
	receipt, _ := rpc.DeployContract(transaction)
	cAddress := receipt.ContractAddress

	logf := NewLogsFilter().AddAddress(cAddress).SetTopic(0, ABI.Events["getHello"].Id())
	wsCli := wsRPC.GetWebSocketClient()
	_, err := wsCli.Subscribe(1, logf, &TestEventHandler{})
	if err != nil {
		t.Error(err.String())
		return
	}

	packed, _ := ABI.Pack("getHello")
	transaction1 := NewTransaction(address).Invoke(cAddress, packed)
	transaction1.Sign(privateKey)
	receipt1, _ := rpc.InvokeContract(transaction1)
	fmt.Println(receipt1.Ret)

	time.Sleep(3 * time.Second)
	wsCli.CloseConn(1)
	time.Sleep(1 * time.Second)
}

func TestWebSocketClient_GetAllSubscription(t *testing.T) {
	bf := NewBlockEventFilter()
	bf.BlockInfo = true
	wsCli := wsRPC.GetWebSocketClient()
	wsCli.CloseConn(1)
	subID, err := wsCli.Subscribe(1, bf, &TestEventHandler{})
	if err != nil {
		t.Error(err.String())
		return
	}

	subs, _ := wsCli.GetAllSubscription(1)
	if len(subs) != 1 {
		t.Errorf("订阅列表长度应该为1，但是得到%d", len(subs))
		return
	}

	err = wsCli.UnSubscribe(subID)
	if err != nil {
		t.Error(err.String())
		return
	}

	subs, _ = wsCli.GetAllSubscription(1)

	if len(subs) != 0 {
		t.Errorf("订阅列表长度应该为0，但是得到%d", len(subs))
	}
}
