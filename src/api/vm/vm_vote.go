package vm

// VoteInit  is for initializing a vote
type VoteInit struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Options     []string `json:"options" binding:"required"`
	SelectType  int      `json:"select_type" des:"1:单选 2:多选"`
	StartTime   int64    `json:"start_time" binding:"required"`
	EndTime     int64    `json:"end_time" binding:"required"`
	CreatorID   uint     `json:"creator_id" binding:"required"`
}

// ChooseOption  is for select one option
type ChooseOption struct {
	VoteID   uint   `json:"vote_id" binding:"required"`
	OptionID uint   `json:"option_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
	UserID   uint   `json:"user_id" binding:"required"`
}

// GetVoteStatus  is for getting status of vote
type GetVoteStatus struct {
	VoteID uint `json:"vote_id" binding:"required"`
	UserID uint `json:"user_id" binding:"required"`
}
