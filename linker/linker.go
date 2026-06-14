package linker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Link(objects []string, output string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("linking is only supported on macOS for now")
	}

	if len(objects) == 0 {
		return fmt.Errorf("link: no object files")
	}

	clang, err := exec.LookPath("clang")
	if err != nil {
		return fmt.Errorf("clang was not found; install Xcode Command Line Tools with `xcode-select --install`")
	}

	args := make([]string, 0, len(objects)+2)
	args = append(args, objects...)
	args = append(args, "-o", output)

	var stderr bytes.Buffer
	cmd := exec.Command(clang, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() != 0 {
			return fmt.Errorf("link failed: %w\n%s", err, stderr.String())
		}
		return fmt.Errorf("link failed: %w", err)
	}

	return nil
}

func LinkObject(object []byte, output string) error {
	tmp, err := os.CreateTemp("", "kaleidoscope-*.o")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(object); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return Link([]string{tmp.Name()}, output)
}
