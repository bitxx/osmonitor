package cmdutil

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

func RunCmd(cmdstring string) (string, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", cmdstring)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("%s", stderr.String()), err
	}
	content := fmt.Sprintf("%v", out.String())
	if content == "" {
		err = errors.New(fmt.Sprintf("command [%s] not found", cmdstring))
		return "", err
	}
	return fmt.Sprintf("%v", out.String()), nil
}
