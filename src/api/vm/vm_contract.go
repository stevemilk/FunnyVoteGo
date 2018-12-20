package vm

// ReqInvokeCon request invoke contract
type ReqInvokeCon struct {
	ContractAddr string `json:"contract_addr"`
	ContractCode string `json:"contract_code"`
	MethodName   string `json:"method_name"`
	MethodParams string `json:"method_params"`
}

//CompileResult compile result of contract
type CompileResult struct {
	Abi   []string `form:"abi" json:"abi"`
	Bin   []string `form:"bin" json:"bin"`
	Types []string `form:"types" json:"types"`
}

//InvokeReturn return after invoke
type InvokeReturn struct {
	Abi       string `json:"abi"`
	Methods   string `json:"methods"`
	Param     string `json:"param"`
	IsSuccess int    `json:"is_success"`
	Result    string `json:"result"`
}

//ReqCompileParam to be compiled info
type ReqCompileParam struct {
	ContractCode string `json:"contract_code" form:"contract_code"`
	ChainID      uint   `json:"chain_id" form:"chain_id"`
}
