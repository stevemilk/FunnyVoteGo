package ssh

import (
	"bytes"
	"fmt"
	"hyperbaas/src/lib/password"
	"hyperbaas/src/util"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/glog"
	"github.com/pkg/sftp"
)

// Client for remote
type Client struct {
	remoteUser string
	remoteIP   string
	remotePort int
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

// NewClient return a new sftp client
func NewClient(user, pwd, host string, port int) (*Client, error) {
	client := &Client{}
	if err := client.connect(user, password.Decode(pwd), host, port); err != nil {
		return nil, err
	}
	client.remoteUser = user
	client.remoteIP = host
	client.remotePort = port
	return client, nil
}

// Close sftp client
func (client *Client) Close() {
	if client.sftpClient != nil {
		client.sftpClient.Close()
	}

	if client.sshClient != nil {
		client.sshClient.Close()
	}
}

// RunCommand run remote linux shell
func (client *Client) RunCommand(shell string) (stdout string, err error) {
	session, err := client.sshClient.NewSession()
	if err != nil {
		return
	}
	defer func() {
		session.Close()
	}()

	var buf bytes.Buffer
	session.Stdout = &buf
	err = session.Run(shell)
	if err != nil {
		return
	}

	stdout = string(buf.Bytes())

	return
}

// NewSession return a ssh-session
func (client *Client) NewSession() (*ssh.Session, error) {
	return client.sshClient.NewSession()
}

// WriteRemoteFile delete old file and create new file
func (client *Client) WriteRemoteFile(data, file string) error {
	_, err := client.LStat(file)
	if err != nil {
		// rename old file
		client.RenameRemoteFile(file, file+".old")
	}

	// create new file
	dstFile, err := client.sftpClient.Create(file)
	_, err = dstFile.Write([]byte(data))
	if err != nil {
		// rollback
		client.RemoteDeleteFile(file)
		client.RenameRemoteFile(file+".old", file)
		return err
	}
	client.RemoteDeleteFile(file + ".old")
	return nil
}

// isDirFile is true when send dir
func (client *Client) send(isDirFile bool, srcFile, desPath, parent string) error {
	psrcFile, err := os.Open(srcFile)
	if err != nil {
		glog.Errorf("open fail: %v", err)
		return err
	}
	defer psrcFile.Close()

	fileLen, _ := psrcFile.Seek(0, os.SEEK_END)
	glog.Infof("file length: %v", fileLen)
	psrcFile.Seek(0, os.SEEK_SET)

	destFilePath := path.Join(desPath, path.Base(srcFile))
	if isDirFile {
		destFilePath = strings.Replace(srcFile, parent, desPath, -1)
		glog.Infof("send file: %v", destFilePath)
	}

	dstFile, err := client.sftpClient.Create(destFilePath)
	if err != nil {
		glog.Errorf("sfttp err: %v", err)
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 1<<23)
	ts := time.Now()
	var sent int64
	glog.Infof("sent percent: %v", sent)
	// TODO: opt read & write
	for {
		n, _ := psrcFile.Read(buf)
		if n == 0 {
			break
		}
		x, err := dstFile.Write(buf[:n])
		if err != nil {
			glog.Error(err)
			return err
		}
		if x != n {
			return fmt.Errorf("write bytes error, offset:%d", sent)
		}
		sent += (int64)(n)
		glog.Info("read: ", n, " write: ", x)
		glog.Info("sent percent: ", sent, (float64)(sent)/(float64)(fileLen))
	}
	glog.Infof("end time sp: %v", time.Now().Sub(ts))

	return client.sftpClient.Chmod(destFilePath, 0755)
}

// RenameRemoteFile will rename file
func (client *Client) RenameRemoteFile(oldname, newname string) error {
	return client.sftpClient.Rename(oldname, newname)
}

// RemoteDeleteFile delete file at remote
func (client *Client) RemoteDeleteFile(remoteFile string) error {
	return client.sftpClient.Remove(remoteFile)
}

// RemoteDeletePath delete remote path
func (client *Client) RemoteDeletePath(remotePath string) error {
	if remotePath == "" || remotePath == "/" {
		return fmt.Errorf("remote path error")
	}
	_, err := client.RunCommand(fmt.Sprintf("rm -rf %s", remotePath))
	return err
}

// RemoteDeleteDir delete dir at remote
func (client *Client) RemoteDeleteDir(remoteDir string) error {
	_, err := client.LStat(remoteDir)
	if err != nil {
		return err
	}
	walker := client.sftpClient.Walk(remoteDir)
	dirs := []string{}
	for {
		if strings.Compare(walker.Path(), "") != 0 {
			if walker.Stat().IsDir() {
				dirs = append(dirs, walker.Path())
			} else {
				// delete file
				err := client.RemoteDeleteFile(walker.Path())
				if err != nil {
					return err
				}
			}
		}
		if !walker.Step() {
			break
		}
	}
	for i := 0; i < len(dirs); i++ {
		// delete empty dir
		err := client.sftpClient.RemoveDirectory(dirs[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// Mkdir make dir for remote
func (client *Client) Mkdir(path string) error {
	return client.sftpClient.Mkdir(path)
}

// LStat path
func (client *Client) LStat(path string) (os.FileInfo, error) {
	return client.sftpClient.Lstat(path)
}

// SendFile to remote
func (client *Client) SendFile(srcFile, desPath string) error {
	// check if file exists
	if inf, err := client.sftpClient.Lstat(path.Join(desPath, path.Base(srcFile))); err == nil && inf != nil {
		return fmt.Errorf("remote %s %s already exists", client.remoteIP, path.Join(desPath, path.Base(srcFile)))
	}

	return client.send(false, srcFile, desPath, "")
}

// SendDir to remote
func (client *Client) SendDir(srcPath, desPath string) error {
	// check if dir exists
	if inf, err := client.sftpClient.Lstat(path.Join(desPath, srcPath)); err == nil && inf != nil {
		return fmt.Errorf("remote %s %s already exists", client.remoteIP, path.Join(desPath, srcPath))
	}

	// parent path
	parent := util.GetParentDirectory(srcPath)

	err := filepath.Walk(srcPath, func(srcPath string, f os.FileInfo, err error) error {
		if f == nil {
			glog.Errorf("unknow error: %v", err)
			return err
		}
		if f.IsDir() {
			mkpath := strings.Replace(srcPath, parent, desPath, -1)
			err = client.sftpClient.Mkdir(mkpath)
			if err != nil {
				glog.Error("mk path: ", mkpath, " fail: ", err)
				return err
			}
			return nil
		}

		return client.send(true, srcPath, desPath, parent)
	})
	if err != nil {
		glog.Errorf("walk error: %v", err)
		return err
	}

	return nil
}

func (client *Client) connect(user, password, host string, port int) error {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		// validate server, we just return nil
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client.sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return err
	}

	// create sftp client
	if client.sftpClient, err = sftp.NewClient(client.sshClient); err != nil {
		return err
	}

	return nil
}

// PullFile pull srcFile from server, save to dstFile
func (client *Client) PullFile(remoteFilePath, localFilePath string) error {
	srcFile, err := client.sftpClient.Open(remoteFilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = srcFile.WriteTo(dstFile); err != nil {
		return err
	}
	return nil
}
