package model

import "github.com/glog"

// Vote model
type Vote struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	SelectType  int      `json:"select_type" des:"1:单选 2:多选"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	CreateTime  string   `json:"create_time"`
	CreatorID   uint     `json:"creator_id"`
	Options     []Option `json:"options"`
	Status      int      `json:"status" des:"1:未开始 2:进行中 3:已结束"`
	UserVoted   int      `json:"user_voted" des:"1:未投票 2:已投票"`
}

// Vote2 model
type Vote2 struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	SelectType     int      `json:"select_type" des:"1:单选 2:多选"`
	StartTime      string   `json:"start_time"`
	EndTime        string   `json:"end_time"`
	CreateTime     string   `json:"create_time"`
	CreatorID      uint     `json:"creator_id"`
	Status         int      `json:"status" des:"1:未开始 2:进行中 3:已结束"`
	UserVoted      int      `json:"user_voted" des:"1:未投票 2:已投票"`
	OptionIDs      []string `json:"option_ids"`
	OptionContents []string `json:"option_contents"`
}

// Option  model
type Option struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Total   uint   `json:"total"`
	VoteID  string `json:"vote_id"`
}

// UserOption  model
type UserOption struct {
	ID            string `json:"id"`
	VoteID        string `json:"vote_id"`
	OptionID      string `json:"option_id"`
	OptionContent string `json:"option_content"`
	UserID        uint   `json:"user_id"`
	Publickey     string `json:"public_key"`
	CreateTime    string `json:"create_time"`
}

// VoteRecord  model
type VoteRecord struct {
	UserID        string `json:"user_id"`
	OptionContent string `json:"option_content"`
	TxHash        string `json:"tx_hash"`
}

// HashRecord  model
type HashRecord struct {
	ID            uint   `json:"id"`
	VoteID        string `json:"vote_id"`
	UserID        uint   `json:"user_id"`
	OptionID      string `json:"option_id"`
	OptionContent string `json:"option_content"`
	TxHash        string `json:"tx_hash"`
}

// CreateHashRecord create hash record
func CreateHashRecord(hr *HashRecord) (*HashRecord, bool) {
	err := db.Create(&hr).Error
	if err != nil {
		glog.Errorf("CreateHashRecord : %v", err)
		return nil, false
	}
	return hr, true
}

// GetHashRecord get hash record
func GetHashRecord(maps interface{}) (*HashRecord, bool) {
	var hr HashRecord
	err := db.Model(&HashRecord{}).Where(maps).Find(&hr).Error
	if err != nil {
		glog.Errorf("GetHashRecord : %v", err)
		return nil, false
	}
	return &hr, true

}
