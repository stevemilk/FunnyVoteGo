package jrpc

import (
	"errors"
	"fmt"
	"github.com/glog"
	"github.com/hyperchain/gosdk/rpc"
	"hyperbaas/src/model"
)

// GetJRPC get availble jrpc
func GetJRPC(node *model.Node) (*rpc.RPC, error) {
	chain, err := model.GetChainByID(node.ChainID)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	path := "./conf/chain_SDK/" + chain.ChainCode + "/conf"
	hpc := rpc.NewRPCWithPath(path)
	if hpc == nil {
		return nil, fmt.Errorf("初始化rpc失败")
	}
	_, err = hpc.GetNodes()
	if err != nil {
		return nil, err
	}
	return hpc, nil
}

// GetJRPCRandom get availble jrpc
func GetJRPCRandom(chain *model.Chain) (int, *rpc.RPC, error) {
	err := errors.New("no availible node")
	for k, v := range chain.Nodes {
		chain, err := model.GetChainByID(v.ChainID)
		if err != nil {
			glog.Error(err)
			return k, nil, err
		}
		path := "./conf/chain_SDK/" + chain.ChainCode + "/conf"
		hpc := rpc.NewRPCWithPath(path)
		if hpc == nil {
			return -2, nil, fmt.Errorf("初始化rpc失败")
		}
		_, err = hpc.GetNodes()
		if hpc != nil && err == nil {
			return k, hpc, nil
		}
	}

	return -1, nil, errors.New("get jrpc fail: " + err.Error())
}
