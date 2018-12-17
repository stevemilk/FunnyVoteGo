package port

import (
	"net"
	"time"
)

// Check return true if port is open
func Check(ip string, port int) (open bool) {
	ch := make(chan bool, 1)
	go func() {
		IP := net.ParseIP(ip)
		PORT := port

		tcpAddr := net.TCPAddr{
			IP:   IP,
			Port: PORT,
		}

		conn, err := net.DialTCP("tcp", nil, &tcpAddr)

		if err == nil {
			ch <- true
			conn.Close()
		} else {
			ch <- false
		}
	}()

	select {
	case lp := <-ch:
		open = lp
		close(ch)
	case <-time.After(2 * time.Second):
		open = false
		close(ch)
	}
	return
}
