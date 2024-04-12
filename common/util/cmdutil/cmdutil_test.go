package cmdutil

import (
	"fmt"
	"testing"
)

func TestCmd(t *testing.T) {
	content, err := RunCmd(fmt.Sprintf("pidof %s", "geth"))
	fmt.Println(err)
	fmt.Println(content)
}
