// package mute implements functions to execute other programs muting std streams if required
package mute

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// EXEC_ERR is exit code when failed to execute the command
const EXEC_ERR = 127

// execContext is the details of an executed command, and the expected conditions to act on
type execContext struct {
	Cmd        string
	ExitCode   int
	StdoutText string
	StderrText string
	Conf       *Conf
	OutWriter  io.Writer
	ErrWriter  io.Writer
}

// Exec runs a command and prints the stdout only when the output matched the configuration
// executes a command, checks the exit codes and matches stdout with patterns,
// and prints the stdout when the cmd spec matched. Return the exit code of cmd.
func Exec(cmd string, args []string, conf *Conf, outWriter io.Writer) (cmdExitCode int) {
	if cmd == "" {
		panic("cmd is empty")
	}
	execCmd := exec.Command(cmd, args...)
	var stdoutBuffer bytes.Buffer
	execCmd.Stdout = &stdoutBuffer
	if err := execCmd.Run(); err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			cmdExitCode = e.ExitCode()
		default:
			cmdExitCode = EXEC_ERR
		}
	}
	ctx := execContext{Cmd: cmd, ExitCode: cmdExitCode, StdoutText: stdoutBuffer.String(), Conf: conf, OutWriter: outWriter, ErrWriter: outWriter}
	crt := cmdCriteria(ctx.Cmd, ctx.Conf)
	if !matchesCriteria(crt, ctx.ExitCode, ctx.StdoutText) {
		fmt.Fprintf(ctx.OutWriter, "%v", ctx.StdoutText)
	}
	return cmdExitCode
}

// matchesCriteria indicates if results of an exec matches a given Criteria
// to decide if a program should be muted or not, its exit code and stdout/stderr is matched
// against the configured Criteria. This function helps to decide on mute or not
func matchesCriteria(criteria *Criteria, code int, stdout string) bool {
	for _, crt := range *criteria {
		if codesContain(crt.ExitCodes, code) {
			return true
		}
	}
	return false
}

// cmdCriteria returns the Criteria that the cmd should be matched against from the Conf
// Each command is matched against a criteria. The Conf has Criterias
// either per command or a default one that is used for all commands.
// cmdCriteria finds the corresponding Criterian from a Conf that the cmd
// should be checked against
func cmdCriteria(cmd string, conf *Conf) *Criteria {
	matched := ""
	for key, _ := range conf.Commands {
		if len(key) > len(matched) && strings.HasPrefix(cmd, key) {
			matched = key
		}
	}
	if matched == "" { // no command specific criteria matched cmd
		return &conf.Default
	}
	criteria := conf.Commands[matched]
	return &criteria
}
