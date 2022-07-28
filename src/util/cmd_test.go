package util

import "testing"

func TestCMD(t *testing.T) {
	CMD("uname", "-a")
	CMD("pwd")
	CMD("whoami")
}
