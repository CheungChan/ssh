# 一个远程执行命令的工具

用法1: 使用本地调用ssh命令方式

```go
package main

import (
	"os"
	"github.com/cheungchan/ssh"
)

func main() {
	var cmdString = "echo 'begin';sleep 3;echo 'end';"
	_ := ssh.RunByExecCmd("jump", cmdString, os.Stdout)
}

```
用法2： 使用包先建立连接，再调用，这种方式可以复用同一个client，做很多操作。
但注意不要开太多，开太多会panic。
```go
package main

import (
	"fmt"
	"github.com/cheungchan/ssh"
	"os"
	"sync"
)

var cmdString = "echo 'begin';sleep 3;echo 'end';"

func main() {
	wg := sync.WaitGroup{}
	config := &ssh.SSHClientConfig{User: "root", Host: "1.1.1.1", Port: "22", PrivateKey: "~/.ssh/id_rsa"}
	client, err := ssh.GetSSHClient(config)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < 10; i++ {
		go func() {
			err := ssh.RunBySSHClient(client, cmdString, os.Stdout)
			if err != nil {
				panic(err)
			}

			wg.Done()
		}()
		wg.Add(1)
	}
	wg.Wait()
}

```