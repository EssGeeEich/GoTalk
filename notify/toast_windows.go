package notify

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func powershellPath() string {
	if root := os.Getenv("SystemRoot"); root != "" {
		path := filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "powershell.exe"
}

// Shows the notification. It blocks until PowerShell exits, which may take a few seconds.
func (n *Notification) Push() error {
	toast, err := n.buildXML()
	if err != nil {
		return err
	}

	appID := n.AppID
	if appID == "" {
		appID = "Windows App"
	}

	cmd := exec.Command(powershellPath(), "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-EncodedCommand", encodeCommand(script))
	cmd.Env = append(os.Environ(), envXML+"="+toast, envAppID+"="+appID)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}
