package email

import "io/ioutil"

var (
	confirm string
)

// InitEmail init config for email`
func InitEmail() {
	cb, err := ioutil.ReadFile("./conf/email/confirm.html")
	if err != nil {
		panic(err)
	}
	confirm = string(cb)
}
