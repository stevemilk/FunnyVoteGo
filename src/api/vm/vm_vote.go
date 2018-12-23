package vm

// VoteInit  is for initializing a vote
type VoteInit struct {
	Title       string   `json:"title" form:"title" binding:"required"`
	Description string   `json:"description" form:"description" binding:"required"`
	Options     []string `json:"options" form:"options" binding:"required"`
	SelectType  int      `json:"select_type" form:"select_type" des:"1:单选 2:多选"`
	StartTime   string   `json:"start_time" form:"start_time" binding:"required"`
	EndTime     string   `json:"end_time" form:"end_time" binding:"required"`
	CreatorID   uint     `json:"creator_id" form:"creator_id" binding:"required"`
}

// ChooseOption  is for select one option
type ChooseOption struct {
	VoteID        string `json:"vote_id" form:"voteid" binding:"required"`
	OptionID      string `json:"option_id" form:"option_id" binding:"required"`
	OptionContent string `json:"option_content" form:"option_content" binding:"required"`
	UserID        uint   `json:"user_id" form:"user_id" binding:"required"`
}

// GetVoteStatus  is for getting status of vote
type GetVoteStatus struct {
	VoteID    string `json:"vote_id" form:"vote_id" binding:"required"`
	UserID    uint   `json:"user_id" form:"user_id" binding:"required"`
	Publickey string `json:"public_bkey" form:"public_key"`
}

// VoteID  model
type VoteID struct {
	VoteID string `json:"vote_id" form:"vote_id" binding:"required"`
}
