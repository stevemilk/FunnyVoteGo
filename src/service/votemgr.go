package service

import (
	"FunnyVoteGo/src/api/vm"

	"github.com/hyperchain/gosdk/utils/ecdsa"
)

func StartVote(voteinit *vm.VoteInit, key *ecdsa.Key) {
	//调用合约新建投票活动
	contractinfo, _ := GetContractInfo("新建投票")
	InvokeContract(vm.ReqInvokeCon{
		ContractAddr: contractinfo.Name,
		ContractCode: contractinfo.Code,
		MethodName:   "新建投票",
		// convert
		MethodParams: "",
	}, key)

}

func AddOptions(options []string, voteid uint) bool {
	//for _, option := range options {
	//调用合约添加选项内容
	//}
	return true
}
