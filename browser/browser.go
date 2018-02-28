package browser

import (
	"os/exec"
	"runtime"
)

func Open(url string) error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}

	prog := args[0]
	args = append(args[1:], url)
	return exec.Command(prog, args...).Start()
}
