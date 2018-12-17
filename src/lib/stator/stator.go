package stator

import (
	"fmt"
	"hyperbaas/src/lib/password"
	"hyperbaas/src/lib/ssh"
	"io"
	"os"
	"strings"

	"github.com/glog"
)

// FmtBytes format bytes info
func FmtBytes(val uint64) string {
	if val < 1024 {
		return fmt.Sprintf("%d bytes", val)
	} else if val < 1024*1024 {
		return fmt.Sprintf("%6.2f KiB", float64(val)/1024.0)
	} else if val < 1024*1024*1024 {
		return fmt.Sprintf("%6.2f MiB", float64(val)/1024.0/1024.0)
	}
	return fmt.Sprintf("%6.2f GiB", float64(val)/1024.0/1024.0/1024.0)
}

//func fmtUptime(stats *Stats) string {
//dur := stats.Uptime
//dur = dur - (dur % time.Second)
//var days int
//for dur.Hours() > 24.0 {
//days++
//dur -= 24 * time.Hour
//}
//s1 := dur.String()
//s2 := ""
//if days > 0 {
//s2 = fmt.Sprintf("%dd ", days)
//}
//for _, ch := range s1 {
//s2 += string(ch)
//if ch == 'h' || ch == 'm' {
//s2 += " "
//}
//}
//return s2
//}

// GetStats return stat info
func GetStats(ip, name, pwd string, port int) (stat *Stats, err error) {
	client, err := ssh.NewClient(name, password.Encode(pwd), ip, port)
	if err != nil {
		return nil, err
	}

	defer func() {
		client.Close()
	}()

	stats := Stats{}
	err = getAllStats(client, &stats)
	if err != nil {
		return nil, err
	}
	stats.CPUStr = fmt.Sprintf("%s X %d", stats.CPU.Version, stats.CPU.Count)
	stats.MemoryStr = strings.TrimSpace(fmt.Sprintf("%s", FmtBytes(stats.MemTotal)))
	if len(stats.FSInfos) > 0 {
		var total uint64
		for _, fs := range stats.FSInfos {
			total += fs.Used
			total += fs.Free
		}
		stats.FSTotal = total
		stats.FSTotalStr = fmt.Sprintf("%s", FmtBytes(stats.FSTotal))
	}
	return &stats, nil
}

func showStats(output io.Writer, client *ssh.Client) {
	stats := Stats{}
	getAllStats(client, &stats)
	fmt.Fprintf(output,
		`%s  OS:
      %s %s %s
 
  CPU:
	  v : %s%s%s, 
	  n : %s%d%s
 
  Memory:
      total    = %s%s%s

  `,
		escClear,
		escBrightWhite, stats.OS, escReset,
		escBrightWhite, stats.CPU.Version, escReset,
		escBrightWhite, stats.CPU.Count, escReset,
		escBrightWhite, FmtBytes(stats.MemTotal), escReset,
	)
	if len(stats.FSInfos) > 0 {
		glog.Info("Filesystems:")
		for _, fs := range stats.FSInfos {
			fmt.Fprintf(output, "    total: %s%s%s point: %s%s%s\n",
				escBrightWhite, FmtBytes(fs.Used+fs.Free), escReset,
				escBrightWhite, fs.MountPoint, escReset,
			)
		}
		fmt.Println()
	}
}

// GetStatInfo start monitor service
func GetStatInfo() {
	client, err := ssh.NewClient("hyperchain", "B1ockch@inhyperchain", "139.219.8.62", 22)
	if err != nil {
		panic(err)
	}
	defer func() {
		client.Close()
	}()

	// test show stat info
	showStats(os.Stdout, client)
}

const (
	escClear       = "\033[H\033[2J"
	escRed         = "\033[31m"
	escReset       = "\033[0m"
	escBrightWhite = "\033[37;1m"
)
