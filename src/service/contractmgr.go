package service

import (
	"FunnyVoteGo/src/api/vm"
	"FunnyVoteGo/src/model"
	"FunnyVoteGo/src/util"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/glog"
	"github.com/hyperchain/gosdk/abi"
	"github.com/hyperchain/gosdk/account"
	"github.com/hyperchain/gosdk/rpc"
	"github.com/hyperchain/gosdk/utils/ecdsa"
	"github.com/tealeg/xlsx"
)

// InitKey create key by key file
func InitKey() (*ecdsa.Key, error) {
	m, err := util.FilePathToMap("./conf/key/key.json")
	if err != nil {
		return nil, err
	}
	k := m["privateKey"].(string)
	key, err := account.NewAccountFromPriv(k)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// InvokeContract invoke contract
func InvokeContract(param vm.ReqInvokeCon, key *ecdsa.Key) (*vm.InvokeReturn, error) {
	cr, err := CompileContract(param.ContractCode)
	if err != nil {
		glog.Error(err)
		return nil, fmt.Errorf("合约编译失败")
	}

	abii := cr.Abi[0]
	ABI, _ := abi.JSON(strings.NewReader(abii))
	var args []interface{}
	if param.MethodParams == "{}" {
		args = nil
	} else {
		args, err = ParseParam(param.MethodParams, ABI.Methods[param.MethodName].Inputs)
		if err != nil {
			glog.Error(err)
			return nil, fmt.Errorf("请上传正确格式参数")
		}
	}
	glog.Info("args is : ", args)
	hpc := rpc.NewRPCWithPath("./conf/chain_SDK/conf")
	if hpc == nil {
		return nil, fmt.Errorf("初始化rpc失败")
	}
	glog.Info("invoke ...")
	packed, err := ABI.Pack(param.MethodName, args...)
	if err != nil {
		glog.Error(err)
		return nil, fmt.Errorf("方法调用失败：调用失败，请检查区块链及合约状态")
	}
	tranInvoke := rpc.NewTransaction(key.GetAddress()).Invoke(param.ContractAddr, packed)
	tranInvoke.Sign(key)
	txInvoke, stdErr := hpc.InvokeContract(tranInvoke)
	if stdErr != nil {
		glog.Error(stdErr)
		return nil, fmt.Errorf("方法调用失败：调用失败，请检查区块链及合约状态")
	}

	//p1, p2, err := constructOutput(ABI, param.MethodName, txInvoke.Ret)

	var result = vm.InvokeReturn{
		Abi:       abii,
		Param:     param.MethodParams,
		IsSuccess: 1,
		Result:    txInvoke.Ret,
		Methods:   param.MethodName,
	}
	return &result, nil
}

// ParseParam parse invoke param
func ParseParam(s string, inputs []abi.Argument) ([]interface{}, error) {
	var args map[string]interface{}
	var param []interface{}
	err := json.Unmarshal([]byte(s), &args)
	if err != nil {
		return nil, err
	}
	for _, v := range inputs {
		param = util.ABIchangeType(param, args[v.Name], v.Type.String())
		//param = append(param, args[v.Name])
	}

	return param, nil
}

//constructOutput construct invoke output
func constructOutput(ABI abi.ABI, MethodName string, Ret string) (int32, string, error) {
	var p1 int32
	var p2 []byte
	res := []interface{}{&p1, &p2}
	if sysErr := ABI.UnpackResult(&res, "insertVote", Ret); sysErr != nil {
		glog.Info(sysErr)
		return 0, "", sysErr
	}
	glog.Info(p1, string(p2))
	return p1, string(p2), nil

	//ol := len(ABI.Methods[MethodName].Outputs)
	//var v []interface{}
	//for i := 0; i < ol; i++ {
	//	var t interface{}
	//	v = append(v, &t)
	//}
	//err := ABI.Unpack(&v, MethodName, common.FromHex(txInvoke.Ret))
	//if err != nil {
	//	glog.Error(err)
	//	return nil, err
	//}

	//return v, nil
	//m := make(map[string]interface{})
	//m[MethodName] = v
	//data, err := json.Marshal(m)
	//if err != nil {
	//	glog.Error(err)
	//	return "", err
	//}
	//return string(data), nil
}

// CompileContract compile contract on the chain
func CompileContract(contractcode string) (*vm.CompileResult, error) {

	path := "./conf/chain_SDK/conf"
	hpc := rpc.NewRPCWithPath(path)
	if hpc == nil {
		return nil, fmt.Errorf("初始化rpc失败")
	}
	res, err := hpc.CompileContract(contractcode)
	if err != nil {
		return nil, fmt.Errorf("合约编译失败")
	}
	var result = vm.CompileResult{
		Abi:   res.Abi,
		Bin:   res.Bin,
		Types: res.Types,
	}
	return &result, nil
}

// GetContractInfo get contract info by contract name
func GetContractInfo(contractname string) (*model.ContractInfo, error) {
	var contractinfo model.ContractInfo
	filename := "conf/contract/contract.xlsx"
	xlFile, err := xlsx.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	for s, sheet := range xlFile.Sheets {
		if s == 1 {
			break
		}
		for r, row := range sheet.Rows {
			if r == 0 {
				continue
			}
			if row.Cells[1].String() == contractname {
				contractinfo.Address = row.Cells[2].String()
				contractinfo.Code = row.Cells[3].String()
				break
			}
		}
	}
	return &contractinfo, nil
}
