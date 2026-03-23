package game

import (
	"os/exec"
	"runtime"
	"strings"
)

func (g *Game) openFilePicker() string {
	if InTestMode {
		return ""
	}
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("osascript", "-e", `POSIX path of (choose file with prompt "Select an .oinakos.yaml save file:" of type {"oinakos.yaml", "yaml"})`)
		out, err := cmd.Output()
		if err != nil { return "" }
		return strings.TrimSpace(string(out))
	}
	return ""
}
