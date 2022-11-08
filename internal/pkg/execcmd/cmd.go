package execcmd

import (
	"bytes"
	"context"
	"go.uber.org/zap"
	"os/exec"
	"strings"
)

type loggerout struct {
	log *zap.SugaredLogger
}

func (l *loggerout) Write(p []byte) (n int, err error) {
	l.log.Infof("%s", bytes.TrimSpace(p))
	return len(p), nil
}

func Run(ctx context.Context, dir string, logger *zap.SugaredLogger, cmd string, args ...string) error {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Dir = dir
	c.Stdout = &loggerout{logger}
	c.Stderr = &loggerout{logger}

	err := c.Run()
	if err != nil {
		if strings.Contains(err.Error(), "signal") {
			return nil
		} else {
			return err
		}
	}
	return nil
}
