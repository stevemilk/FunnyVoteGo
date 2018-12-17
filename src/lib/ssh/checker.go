package ssh

import "hyperbaas/src/lib/password"

// CheckSSH return true when ssh is ok
func CheckSSH(ip, username, pwd string, port int) (bool, error) {
	client, err := NewClient(username, password.Encode(pwd), ip, port)
	defer func() {
		client.Close()
	}()
	return client != nil && err == nil, err
}
