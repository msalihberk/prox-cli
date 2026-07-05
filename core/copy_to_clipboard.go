/* Copyright 2026 Mustafa Salih Berk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package core

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"runtime"
)

func CopyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("clip")
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	default:
		return errors.New("unsupported operating system for clipboard operations")
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			if runtime.GOOS == "linux" {
				return errors.New("clipboard utility 'xclip' not found. Please install it using your package manager (e.g., apt install xclip)")
			}
			return errors.New("required system clipboard utility is missing")
		}
		return err
	}

	if _, err := io.WriteString(in, text); err != nil {
		return err
	}
	in.Close()

	if err := cmd.Wait(); err != nil {
		if stderr.Len() > 0 {
			return errors.New(stderr.String())
		}
		return err
	}

	return nil
}
