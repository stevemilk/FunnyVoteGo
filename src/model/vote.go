package model

// Vote model
type Vote struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SelectType  int    `json:"select_type" des:"1:单选 2:多选"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	CreateTime  int64  `json:"create_time"`
	CreatorID   uint   `json:"creator_id"`
}

// Option  model
type Option struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Total   uint   `json:"total"`
	VoteID  uint   `json:"vote_id"`
}

// UserOption  model
type UserOption struct {
	ID       string `json:"id"`
	VoteID   uint   `json:"vote_id"`
	OptionID uint   `json:"option_id"`
	Content  string `json:"content"`
	UserID   uint   `json:"user_id"`
}
