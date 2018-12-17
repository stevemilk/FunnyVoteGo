package ssh

import (
	"fmt"
	"hyperbaas/src/lib/password"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	pwd = password.Encode("B1ockch@in")
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("root", pwd, "172.16.7.101", 22)
	defer func() {
		client.Close()
	}()

	Convey("TestNewClient", t, func() {
		Convey("create sshclient should not return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("sshclient should not be nil", func() {
			So(client, ShouldNotBeNil)
		})

		Convey("sshclient ip should be 172.16.7.101", func() {
			So(client.remoteIP, ShouldEqual, "172.16.7.101")
		})

		Convey("sshclient username should be root", func() {
			So(client.remoteUser, ShouldEqual, "root")
		})

		Convey("sshclient port should be 22", func() {
			So(client.remotePort, ShouldEqual, 22)
		})

		Convey("client.sshclient should not be nil", func() {
			So(client.sshClient, ShouldNotBeNil)
		})

		Convey("sftpClient should not be nil", func() {
			So(client.sftpClient, ShouldNotBeNil)
		})
	})
}

func TestSendFile(t *testing.T) {

	client, err := NewClient("root", pwd, "172.16.7.101", 22)
	defer func() {
		client.Close()
	}()

	pwd, _ := os.Getwd()

	Convey("TestSendFile", t, func() {
		Convey("create sshclient should not return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("sshclient send file return nil", func() {
			So(client.SendFile(pwd+"/shell_test.sh", "/root/apanoo"), ShouldBeNil)
		})

		fi, err := client.LStat("/root/apanoo/shell_test.sh")
		Convey("sshclient send file success", func() {
			So(fi, ShouldNotBeNil)
		})

		Convey("sshclient send file success should not return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("sshclient sent file name should be 'shell_test.sh'", func() {
			So(fi.Name(), ShouldEqual, "shell_test.sh")
		})
	})

	//client, err := NewClient("root", "blockchain", "172.16.7.101", 22)
	//defer func() {
	//	client.Close()
	//}()
	//
	//pwd, _ := os.Getwd()
	//
	//Convey("TestSendFile", t, func() {
	//	Convey("create sshclient should not return error", func() {
	//		So(err, ShouldBeNil)
	//	})
	//
	//	Convey("sshclient send file return nil", func() {
	//		So(client.SendFile(pwd+"/node_running.sh", "/root/apanoo"), ShouldBeNil)
	//	})
	//
	//	fi, err := client.LStat("/root/apanoo/node_running.sh")
	//	Convey("sshclient send file success", func() {
	//		So(fi, ShouldNotBeNil)
	//	})
	//
	//	Convey("sshclient send file success should not return error", func() {
	//		So(err, ShouldBeNil)
	//	})
	//
	//	Convey("sshclient sent file name should be 'node_running.sh'", func() {
	//		So(fi.Name(), ShouldEqual, "node_running.sh")
	//	})
	//})
}

func TestRunShell(t *testing.T) {
	client, err := NewClient("root", pwd, "172.16.7.101", 22)
	desShellPath := "/root/apanoo"
	sellName := "shell_test.sh"
	runShell := fmt.Sprintf("cd %s; sh %s/%s", desShellPath, desShellPath, sellName)
	createFile := fmt.Sprintf("%s/sshclient_shell_test.test", desShellPath)
	defer func() {
		client.Close()
	}()

	_, runErr := client.RunCommand(runShell)

	Convey("TestRunShell", t, func() {

		Convey("create sshclient should not return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("sshclient run shell should not return error", func() {
			So(runErr, ShouldBeNil)
		})

		fi, err := client.LStat(createFile)

		Convey("sshclient run shell create file success", func() {
			So(err, ShouldBeNil)
		})

		Convey("sshclient run shell create file get file info success", func() {
			So(fi, ShouldNotBeNil)
		})

		Convey("sshclient run shell create file name should be 'sshclient_shell_test.test'", func() {
			So(fi.Name(), ShouldEqual, "sshclient_shell_test.test")
		})
	})

	//client, err := NewClient("root", "blockchain", "172.16.7.101", 22)
	//desShellPath := "/root/apanoo"
	//shellName := "node_running.sh"
	//runShell := fmt.Sprintf("cd %s; sh %s/%s", desShellPath, desShellPath, shellName)
	//defer func() {
	//	client.Close()
	//}()
	//
	//_, runErr := client.RunCommand(runShell)
	//
	//Convey("TestRunShell", t, func() {
	//
	//	Convey("create sshclient should not return error", func() {
	//		So(err, ShouldBeNil)
	//	})
	//
	//	Convey("sshclient run shell should not return error", func() {
	//		So(runErr, ShouldBeNil)
	//	})
	//
	//})
}

func TestRemovefile(t *testing.T) {
	desShellPath := "/root/apanoo"
	createFile := fmt.Sprintf("%s/sshclient_shell_test.test", desShellPath)

	client, err := NewClient("root", pwd, "172.16.7.101", 22)
	defer func() {
		client.Close()
	}()

	Convey("TestRemoveFile", t, func() {

		Convey("create sshclient should not return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("sshclient remove exit file success", func() {
			So(client.RemoteDeleteFile("/root/apanoo/shell_test.sh"), ShouldBeNil)
		})

		Convey("sshclient failed to delete non-existing file", func() {
			So(client.RemoteDeleteFile("/root/apanoo/shell_test.sh"), ShouldNotBeNil)
		})

		Convey("sshclient remove create file success", func() {
			So(client.RemoteDeleteFile(createFile), ShouldBeNil)
		})

		Convey("sshclient failed to delete non-existing file which create by shell", func() {
			So(client.RemoteDeleteFile(createFile), ShouldNotBeNil)
		})
	})
}

func TestSendDir(t *testing.T) {

	client, err := NewClient("root", pwd, "172.16.7.101", 22)
	defer func() {
		client.Close()
	}()

	pwd, _ := os.Getwd()

	Convey("TestSendDir", t, func() {
		Convey("create sshclient should not return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("sshclient send file return nil", func() {
			So(client.SendDir(pwd, "/root/apanoo"), ShouldBeNil)
		})

		fi, err := client.LStat("/root/apanoo/ssh")
		Convey("sshclient send path success", func() {
			So(fi, ShouldNotBeNil)
		})

		Convey("sshclient send path success should not return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("sshclient sent path name should be 'ssh'", func() {
			So(fi.Name(), ShouldEqual, "ssh")
		})

		Convey("remove sent path should success", func() {
			So(client.RemoteDeleteDir("/root/apanoo/ssh"), ShouldBeNil)
		})

		Convey("remove sent path again should failed", func() {
			So(client.RemoteDeleteDir("/root/apanoo/ssh"), ShouldNotBeNil)
		})
	})
}

func TestWriteRemoteFile(t *testing.T) {
	client, err := NewClient("root", pwd, "172.16.7.101", 22)
	defer func() {
		client.Close()
	}()

	er := client.WriteRemoteFile("test string", "/root/apanoo/1111111.txt")

	e := client.RemoteDeleteFile("/root/apanoo/1111111.txt")

	Convey("TestWriteRemoteFile", t, func() {
		Convey("create sshclient should not return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("write remote file should not return error", func() {
			So(er, ShouldBeNil)
		})

		Convey("delete remote file should not return error", func() {
			So(e, ShouldBeNil)
		})
	})
}

func TestPullFile(t *testing.T) {
	password := password.Encode("B1ockch@in")
	remoteFilePath := "/root/test.sh"
	client, err := NewClient("root", password, "172.16.7.103", 22)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer func() {
		client.Close()
	}()
	pwd, _ := os.Getwd()

	Convey("TestPullFile", t, func() {
		Convey("create sshclient should not return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("sshclient pull file return nil", func() {
			So(client.PullFile(remoteFilePath, pwd+"/backup.txt"), ShouldBeNil)
		})

		fi, err := os.Lstat(pwd + "/backup.txt")
		Convey("sshclient pull file success", func() {
			So(fi, ShouldNotBeNil)
		})

		Convey("sshclient pull file success should not return error", func() {
			So(err, ShouldBeNil)
		})

		Convey("sshclient pull file name should be 'backup.txt'", func() {
			So(fi.Name(), ShouldEqual, "backup.txt")
		})

		Convey("sshclient delete file should return nil", func() {
			So(os.Remove(pwd+"/backup.txt"), ShouldBeNil)
		})
	})
}
