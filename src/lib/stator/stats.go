package stator

import (
	"bufio"
	"hyperbaas/src/lib/ssh"
	"strconv"
	"strings"
)

// FSInfo is the file system info
type FSInfo struct {
	MountPoint string `json:"mount_point"`
	Used       uint64 `json:"used"`
	Free       uint64 `json:"free"`
}

// CPUInfo is the cpu info
type CPUInfo struct {
	Version string `json:"version"`
	Count   int    `json:"count"`
}

// Stats model
type Stats struct {
	//Uptime     time.Duration
	Hostname   string   `json:"hostname"`
	OS         string   `json:"os"`
	MemTotal   uint64   `json:"mem_total"`
	FSInfos    []FSInfo `json:"fs_info"`
	FSTotal    uint64   `json:"fs_total"`
	CPU        CPUInfo  `json:"cpu_info"` // or []CPUInfo to get all the cpu-core's stats?
	CPUStr     string   `json:"cup_str"`
	FSTotalStr string   `json:"fs_total_str"`
	MemoryStr  string   `json:"memory_str"`
}

func getHostname(client *ssh.Client, stats *Stats) (err error) {
	hostname, err := client.RunCommand("/bin/hostname -f")
	if err != nil {
		return
	}

	stats.Hostname = strings.TrimSpace(hostname)
	return
}

func getMemInfo(client *ssh.Client, stats *Stats) (err error) {
	lines, err := client.RunCommand("/bin/cat /proc/meminfo")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 3 {
			val, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				continue
			}
			val *= 1024
			switch parts[0] {
			case "MemTotal:":
				stats.MemTotal = val
			}
		}
	}

	return
}

func getFSInfo(client *ssh.Client, stats *Stats) (err error) {
	lines, err := client.RunCommand("/bin/df -B1")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	flag := 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		n := len(parts)
		dev := n > 0 && strings.Index(parts[0], "/dev/") == 0
		if n == 1 && dev {
			flag = 1
		} else if (n == 5 && flag == 1) || (n == 6 && dev) {
			i := flag
			flag = 0
			used, err := strconv.ParseUint(parts[2-i], 10, 64)
			if err != nil {
				continue
			}

			free, err := strconv.ParseUint(parts[3-i], 10, 64)
			if err != nil {
				continue
			}
			stats.FSInfos = append(stats.FSInfos, FSInfo{
				parts[5-i], used, free,
			})
		}
	}

	return
}

func getCPU(client *ssh.Client, stats *Stats) (err error) {
	lines, err := client.RunCommand("/bin/cat /proc/cpuinfo | grep name")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	var vs string
	for scanner.Scan() {
		line := scanner.Text()
		spls := strings.Split(line, ":")
		if len(spls) == 2 {
			vs = spls[1]
			stats.CPU.Count = len(spls)
		}
	}
	stats.CPU.Version = strings.TrimLeft(vs, " ")

	return
}

func getOS(client *ssh.Client, stats *Stats) (err error) {

	lines, err := client.RunCommand("/bin/cat /etc/os-release")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	var final string
	for scanner.Scan() {
		line := scanner.Text()
		spls := strings.Split(line, "=")
		if len(spls) == 2 {
			if spls[0] == "NAME" {
				final = spls[1]

			}
			if spls[0] == "PRETTY_NAME" {
				stats.OS = strings.Replace(spls[1], "\"", "", -1)
			}
		}
	}
	if stats.OS == "" {
		stats.OS = strings.Replace(final, "\"", "", -1)
	}

	return
}

func getAllStats(client *ssh.Client, stats *Stats) (err error) {
	//getUptime(client, stats)
	err = getHostname(client, stats)
	if err != nil {
		return
	}
	err = getMemInfo(client, stats)
	if err != nil {
		return
	}
	err = getFSInfo(client, stats)
	if err != nil {
		return
	}
	err = getCPU(client, stats)
	if err != nil {
		return
	}
	err = getOS(client, stats)
	if err != nil {
		return
	}
	return
}
