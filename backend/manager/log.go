package manager

import (
	"errors"
	"os/exec"

	"case/backend/pkg/logger"
)

func logErrorOutput(err error) {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		logger.Errorf("command stderr: %v", string(exitErr.Stderr))
	}
}
