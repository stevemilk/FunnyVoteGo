package model

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
