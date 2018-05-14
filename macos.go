// +build darwin

package clipboard

import (
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

func Get() (string, error) {
	buf, err := exec.Command("pbpaste", "-Prefer", "txt").Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to run pbpaste command")
	}

	return string(buf), nil
}

func Set(value string) error {
	cmd := exec.Command("pbcopy")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "failed to get stdin for pbcopy")
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, value)
	}()

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to run pbcopy command")
	}

	return nil
}
