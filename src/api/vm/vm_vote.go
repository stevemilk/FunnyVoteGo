package vm

// VoteInit  is for initializing a vote
type VoteInit struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Options     []string `json:"options" binding:"required"`
	SelectType  int      `json:"select_type" des:"1:单选 2:多选"`
	StartTime   string   `json:"start_time" binding:"required"`
	EndTime     string   `json:"end_time" binding:"required"`
	CreatorID   uint     `json:"creator_id" binding:"required"`
}

// ChooseOption  is for select one option
type ChooseOption struct {
	VoteID        string `json:"vote_id" binding:"required"`
	OptionID      string `json:"option_id" binding:"required"`
	OptionContent string `json:"option_content" binding:"required"`
	UserID        uint   `json:"user_id" binding:"required"`
}

// GetVoteStatus  is for getting status of vote
type GetVoteStatus struct {
	VoteID    string `json:"vote_id" binding:"required"`
	UserID    uint   `json:"user_id" binding:"required"`
	Publickey string `json:"public_bkey"`
}
