//go:build !windows

package tui

import (
	"io/ioutil"
	"os"
	"syscall"
)

var consoleDevice string = "/dev/tty"
var devPrefixes = [...]string{"/dev/pts/", "/dev/"}

func ttyname() string {
	var stderr syscall.Stat_t
	if syscall.Fstat(2, &stderr) != nil {
		return ""
	}

	for _, prefix := range devPrefixes {
		files, err := ioutil.ReadDir(prefix)
		if err != nil {
			continue
		}

		for _, file := range files {
			if stat, ok := file.Sys().(*syscall.Stat_t); ok && stat.Rdev == stderr.Rdev {
				return prefix + file.Name()
			}
		}
	}
	return ""
}

// TtyIn returns terminal device to be used as STDIN, falls back to os.Stdin
func TtyIn() *os.File {
	in, err := os.OpenFile(consoleDevice, syscall.O_RDWR, 0)
	if err != nil {
		tty := ttyname()
		if len(tty) > 0 {
			if in, err := os.OpenFile(tty, syscall.O_RDWR, 0); err == nil {
				return in
			}
		}
		return os.Stdin
	}
	return in
}
