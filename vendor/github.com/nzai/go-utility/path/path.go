package path

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetStartupDir 获取起始目录
func GetStartupDir() (string, error) {

	if strings.HasPrefix(os.Args[0], "./") {
		return os.Getwd()
	}

	startupPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	//    Is Symlink
	fi, err := os.Lstat(startupPath)
	if err != nil {
		return "", err
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		startupPath, err = os.Readlink(startupPath)
		if err != nil {
			return "", err
		}
	}
	startupDir := filepath.Dir(startupPath)
	if startupDir == "." {
		startupDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	return startupDir, nil
}
