package licensor

import (
	"bytes"
	"fmt"
	"github.com/glog"
	"github.com/gobuild/log"
	"os"
	"os/exec"
	"strings"
)

// GenerateLicensor generate license
func GenerateLicensor(month int, ip string) (path string, err error) {

	log.Infof("begin to generate license, month is %d, ip is %v", month, ip)

	pwd, _ := os.Getwd()

	command := fmt.Sprintf("%s%s%s", pwd, "/tools/licensor/", "licensor")

	if ip != "" {
		command = fmt.Sprintf("%s%s%s%s%d", command, " -b ", ip, " -m ", month)
	} else {
		command = fmt.Sprintf("%s%s%d", command, " -u -m ", month)
	}

	glog.Infof("licensor command: %v", command)

	output, err := execShell(command)

	output = strings.TrimSpace(output)

	glog.Info(output)

	if index := strings.Index(output, "path: "); index != -1 {
		expIndex := strings.Index(output, "EXP_")
		licensepath := output[index+6 : expIndex+12]
		path = fmt.Sprintf("%s%s%s", pwd, "/", licensepath)
		glog.Infof("license path: %v", path)
	}

	glog.Info("end to generate license")

	return

}

func execShell(s string) (output string, err error) {

	cmd := exec.Command("/bin/bash", "-c", s)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		glog.Error("rum command err: ", err)
		return
	}
	output = out.String()
	glog.Infof("run licensor command output: %v", output)
	return
}
