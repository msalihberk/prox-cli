//go:build !windows

package core

import "os"

func replaceFile(source, target string) error {
	return os.Rename(source, target)
}
