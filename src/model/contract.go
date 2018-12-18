package model

// ContractInfo model
type ContractInfo struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Code    string `json:"code"`
}
