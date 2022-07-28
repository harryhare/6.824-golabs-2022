package util

import "os/exec"

func CMD(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	str, _ := cmd.Output()
	Println(string(str))
}
