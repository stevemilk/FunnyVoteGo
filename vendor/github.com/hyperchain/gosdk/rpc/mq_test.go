package rpc

import (
	"fmt"
	"github.com/hyperchain/gosdk/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MQListener struct {
	Message string
}

func (ml *MQListener) HandleDelivery(data []byte) {
	ml.Message = string(data)
	fmt.Println(ml.Message)
}

func TestMqClient_Register(t *testing.T) {
	t.Skip("mq not exist")
	client := rpc.GetMqClient()
	_, err := client.InformNormal(1, "")
	assert.Equal(t, true, err == nil)
	var hash common.Hash
	hash.SetString("123")
	rm := NewRegisterMeta(guomiKey.GetAddress(), "node1queue1", MQBlock).SetTopics(1, hash)
	rm.Sign(guomiKey)
	regist, err := client.Register(1, rm)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, regist.QueueName, "node1queue1")

	//listener := &MQListener{}
	//client.Listen(listener)
}

func TestMqClient_UnRegister(t *testing.T) {
	t.Skip("mq not exist")
	client := rpc.GetMqClient()
	meta := NewUnRegisterMeta(guomiKey.GetAddress(), "node1queue1", "global_fa34664e_1541655230749576905")
	meta.Sign(guomiKey)
	unRegist, err := client.UnRegister(1, meta)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, true, unRegist.Success)
}

func TestMqClient_GetAllQueueNames(t *testing.T) {
	t.Skip("mq not exist")
	client := rpc.GetMqClient()
	queues, err := client.GetAllQueueNames(1)
	if err != nil {
		t.Error(err)
		return
	}
	for _, val := range queues {
		fmt.Println(val)
	}
}
