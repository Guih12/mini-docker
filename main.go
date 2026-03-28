package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	cmd := exec.Command("/proc/self/exe", "child")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if len(os.Args) > 1 && os.Args[1] == "child" {
		fmt.Println("dentro do child 🚀")

		syscall.Sethostname([]byte("container"))

		if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
			fmt.Println("erro mount:", err)
		}

		sh := exec.Command("/bin/sh")
		sh.Stdin = os.Stdin
		sh.Stdout = os.Stdout
		sh.Stderr = os.Stderr

		sh.Run()

		syscall.Unmount("/proc", 0)

		return
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS,
	}

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}
