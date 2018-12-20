package util

import (
	"strconv"
	"time"
)

// GetNowTimeString get now time
func GetNowTimeString() string {
	time := time.Now().Unix()
	timestr := strconv.Itoa(int(time))
	return timestr
}
