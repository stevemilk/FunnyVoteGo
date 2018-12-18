package vm

// VoteInit  is for initializing a vote
type VoteInit struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Options     []string `json:"options" binding:"required"`
	SelectType  int      `json:"select_type" des:"1:单选 2:多选"`
	StartTime   int64    `json:"start_time" binding:"required"`
	EndTime     int64    `json:"end_time" binding:"required"`
}
