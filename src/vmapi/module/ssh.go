package module

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"time"
)

func SSHConnect(host, username, password string, port int) *ssh.Client {
	conf := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	sshClient, _ := ssh.Dial("tcp", addr, conf)
	//defer sshClient.Close()
	return sshClient
}

func SSHExec(host, username, password string, port int, command string) error {
	sshClient := SSHConnect(host, username, password, port)
	defer sshClient.Close()
	// 创建ssh-session
	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var stdOut, stdErr bytes.Buffer
	session.Stdout = &stdOut
	session.Stderr = &stdErr

	session.Run(command)
	if stdErr.String() != "" {
		// log.Fatal("err: ", stdErr.String())
		return errors.New(stdErr.String())
	}
	//log.Println(stdOut.String())
	return nil
}

func SFTPut(host, username, password string, port int, src, dest string)  error {
	sshClient := SSHConnect(host, username, password, port)
	defer sshClient.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	// dest 文件
	destFile, err := sftpClient.Create(dest)
	if err != nil {
		log.Fatal(err)
	}
	defer destFile.Close()

	srcdata := []byte(src)
	if err != nil {
		return err
	}
	destFile.Write(srcdata)
	return nil
}
