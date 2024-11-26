package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type Output struct {
	Stdout, Stderr []byte
	Code           int
	Err            error
}

func Run(ctx context.Context, cmd CommandArgsEnv) (out Output) {
	c := exec.CommandContext(ctx, cmd.Command(), cmd.Args()...)
	c.Env = cmd.Environment()
	var stdout, stderr bytes.Buffer
	c.Stdout, c.Stderr = &stdout, &stderr
	err := c.Run()
	out.Stdout = stdout.Bytes()
	out.Stderr = stderr.Bytes()
	out.Code = c.ProcessState.ExitCode()
	out.Err = fmt.Errorf("code(%d) stderr(%q) error(%v)", out.Code, out.Stderr, err)
	return out
}

func (out *Output) Error() error {
	if out.Err != nil {
		return out.Err
	} else if out.Code != 0 {
		return fmt.Errorf("exit code(%d)", out.Code)
	}
	return nil
}
