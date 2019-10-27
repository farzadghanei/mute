package mute

import (
	"bufio"
	"bytes"
	"testing"
)

func TestExec(t *testing.T) {
	conf := new(Conf)
	var buf bytes.Buffer
	var bufWriter = bufio.NewWriter(&buf)
	got := Exec("go", []string{"version"}, conf, bufWriter)
	if got != 0 {
		t.Errorf("Exec return val. got: %d want: 0", got)
	}
}
