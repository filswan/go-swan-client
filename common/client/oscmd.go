package client

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/filswan/go-swan-client/logs"
)

const SHELL_TO_USE = "bash"

func ExecOsCmd2Screen(cmdStr string, checkStdErr bool) (string, error) {
	out, err := ExecOsCmdBase(cmdStr, true, checkStdErr)
	return out, err
}

func ExecOsCmd(cmdStr string, checkStdErr bool) (string, error) {
	out, err := ExecOsCmdBase(cmdStr, false, checkStdErr)
	return out, err
}

func ExecOsCmdBase(cmdStr string, out2Screen bool, checkStdErr bool) (string, error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd := exec.Command(SHELL_TO_USE, "-c", cmdStr)

	if out2Screen {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	} else {
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
	}

	err := cmd.Run()
	if err != nil {
		logs.GetLogger().Error(cmdStr)
		logs.GetLogger().Error(stderrBuf.String())
		logs.GetLogger().Error(err)
		return "", err
	}

	if checkStdErr {
		if len(stderrBuf.String()) != 0 {
			outErr := errors.New(stderrBuf.String())
			logs.GetLogger().Error(cmdStr)
			logs.GetLogger().Error(outErr)
			return "", outErr
		}
	}

	return stdoutBuf.String(), nil
}
