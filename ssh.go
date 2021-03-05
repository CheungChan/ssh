package ssh

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
)

type SSHClientConfig struct {
	Host       string
	Port       string
	PrivateKey string
	User       string
	Password   string
}

func RunByExecCmd(host string, cmdString string, out io.Writer) error {
	/*
		通过调用本地ssh命令，并实时获得输出。输出会写入out里面。等命令执行完，函数返回
	*/
	cmd := exec.Command("ssh", host, cmdString)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		// 写入out
		_, err = fmt.Fprint(out, line)
		if err != nil {
			return err
		}
	}
	err = cmd.Wait()
	return err
}
func RunBySSHClient(client *ssh.Client, cmd string, out io.Writer) error {
	/*
		通过在client上执行cmd，并实时获得输出。输出会写入out里面。等命令执行完，函数返回
	*/
	// Create a session. It is one session per command.
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	stdout, err := session.StdoutPipe()
	err = session.Start(cmd)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		// 写入out
		_, err = fmt.Fprint(out, line)
		if err != nil {
			return err
		}
	}
	err = session.Wait()
	return err
}
func GetSSHClient(c *SSHClientConfig) (*ssh.Client, error) {
	/*
		根据配置获得client
	*/
	var config *ssh.ClientConfig
	if c.PrivateKey != "" {
		config = &ssh.ClientConfig{
			User: c.User,
			Auth: []ssh.AuthMethod{
				publicKeyAuthFunc(c.PrivateKey),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
		}
	} else if c.Password != "" {
		config = &ssh.ClientConfig{
			User: c.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(c.Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
		}
	}else{
		return nil, errors.New("密码或秘钥必须设置一个")
	}

	// Connect
	client, err := ssh.Dial("tcp", net.JoinHostPort(c.Host, c.Port), config)
	if err != nil {
		return nil, err
	}
	return client, err
}

func publicKeyAuthFunc(kPath string) ssh.AuthMethod {
	keyPath, err := homedir.Expand(kPath)
	if err != nil {
		log.Fatal("find key's home dir failed", err)
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatal("ssh key file read failed", err)
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("ssh key signer failed", err)
	}
	return ssh.PublicKeys(signer)
}
