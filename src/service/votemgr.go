package service

import (
	"FunnyVoteGo/src/api/vm"
	"FunnyVoteGo/src/model"
	"FunnyVoteGo/src/util"
	"io/ioutil"
	"time"

	"github.com/glog"
	"github.com/hyperchain/gosdk/utils/ecdsa"
)

const (
	ContractAddress = "0x490653959952058b2302958152c139ae7fe889e3"
)

// GetContractCode get contract code
func GetContractCode() string {
	cbyte, err := ioutil.ReadFile("./conf/contract/vote.sol")

	if err != nil {
		return ""
	}
	contractcode := string(cbyte[:])
	return contractcode

}

// StartVote start  a vote
func StartVote(voteinit *vm.VoteInit) (uint, bool) {
	//调用合约新建投票活动
	// get contract addr and Code
	contractcode := GetContractCode()
	// init params
	vote := model.Vote{
		ID:          util.StringUUID(),
		Title:       voteinit.Title,
		Description: voteinit.Description,
		SelectType:  voteinit.SelectType,
		StartTime:   voteinit.StartTime,
		EndTime:     voteinit.EndTime,
		CreateTime:  time.Now().Unix(),
		CreatorID:   voteinit.CreatorID,
	}
	params := util.Struct2String(vote)
	if params == "" {

		glog.Info("222")
		return 0, false
	}
	key, err := InitKey()
	if err != nil {
		glog.Info("333")
		return 0, false
	}

	ret, err := InvokeContract(vm.ReqInvokeCon{
		ContractAddr: ContractAddress,
		ContractCode: contractcode,
		MethodName:   "insertVote",
		MethodParams: params,
	}, key)
	if err != nil {
		glog.Info("444j")
		return 0, false
	}

	glog.Info(ret)
	//  TODO: handle with output

	//b := AddOptions(voteinit.Options, 1, key)
	//if !b {
	//	return 0, false
	//}
	return 1, true

}

// AddOptions add options for a vote
func AddOptions(options []string, voteid uint, key *ecdsa.Key) bool {
	for _, option := range options {
		//调用合约添加选项内容
		params := util.Struct2String(model.Option{
			Content: option,
			VoteID:  voteid,
		})

		contractcode := GetContractCode()
		_, err := InvokeContract(vm.ReqInvokeCon{
			ContractAddr: ContractAddress,
			ContractCode: contractcode,
			MethodName:   "insertVoteDetail",
			MethodParams: params,
		}, key)
		if err != nil {
			return false
		}

	}
	return true
}

func ChooseOption(chooseoption *vm.ChooseOption) bool {
	contractcode := GetContractCode()
	key, err := InitKey()
	if err != nil {
		glog.Info("333")
		return false
	}

	// 选项+1
	params1 := util.Struct2String(model.Option{
		VoteID: chooseoption.VoteID,
	})
	ret1, err := InvokeContract(vm.ReqInvokeCon{
		ContractAddr: ContractAddress,
		ContractCode: contractcode,
		MethodName:   "选项+1",
		MethodParams: params1,
	}, key)
	if err != nil {
		return false
	}

	glog.Info(ret1)
	// 插入用户id,选项id等
	params2 := util.Struct2String(model.UserOption{
		ID:       util.StringUUID(),
		VoteID:   chooseoption.VoteID,
		OptionID: chooseoption.OptionID,
		Content:  chooseoption.Content,
		UserID:   chooseoption.UserID,
	})
	ret2, err := InvokeContract(vm.ReqInvokeCon{
		ContractAddr: ContractAddress,
		ContractCode: contractcode,
		MethodName:   "用户选项信息上链",
		MethodParams: params2,
	}, key)
	if err != nil {
		return false
	}

	glog.Info(ret2)
	return true

}

func GetVoteStatus(getvotestatus *vm.GetVoteStatus) {
	// 判断投票在什么时间段

	// 判断用户是否已经投票

	// 返回投票活动信息，每个选项的票数

	// 返回账户id及选项
}
