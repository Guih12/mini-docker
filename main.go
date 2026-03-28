package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const (
	LINUX_ALPINE_PATH  = "/tmp/alpine"
	PROC_PATH          = "/proc"
	PROC_SELF_EXEC     = "/proc/self/exe"
	CONTAINER_HOSTNAME = "my-container"
)

// a principal ideia aqui é, que processo pai executa, cria os namespaces e lança
// o binario como seu filho e o filho ja roda dentro dos namespaces, e faz todas as configuracoes
// como mudar o hostname, mountar o proc que é a arvore de processos daquele container e executa um shell
func changeChroot(newRoot string, pultOld string) {
	if err := syscall.PivotRoot(newRoot, pultOld); err != nil {
		fmt.Println("error on change chroor", err)
	}
}

func bindMount(source string) {
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		fmt.Println("error on mount", err)
	}

	if err := syscall.Mount(source, source, "", syscall.MS_BIND, ""); err != nil {
		fmt.Println("error", err)
	}
}

func main() {
	cmd := exec.Command(PROC_SELF_EXEC, "child")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if len(os.Args) > 1 && os.Args[1] == "child" {
		fmt.Println("dentro do child 🚀")

		syscall.Sethostname([]byte(CONTAINER_HOSTNAME))
		bindMount(LINUX_ALPINE_PATH)

		os.MkdirAll("/tmp/alpine/old-root", 0o700)
		syscall.Chdir(LINUX_ALPINE_PATH)

		changeChroot(LINUX_ALPINE_PATH, "/tmp/alpine/old-root")

		syscall.Chdir("/")
		if err := syscall.Unmount("old-root", syscall.MNT_DETACH); err != nil {
			fmt.Println("error unmount", err)
		}

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
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}
