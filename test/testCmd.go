package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("ping", "www.baidu.com")
	cmd.Start()
	fmt.Println(cmd.Process.Pid) // 打印进程pid
	//cmd.Process.Kill() // 你要的结束进程
	//cmd.Process.Signal(syscall.SIGINT) // 给进程发送信号
	cmd.Wait()
}
