// mute executes programs suppressing std streams if required
// license: MIT, see LICENSE for details.
package main

import (
	"fmt"
	"os"

	"github.com/farzadghanei/mute"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Version %v. Usage: %v COMMAND\n", mute.Version, os.Args[0])
		os.Exit(mute.ExitErrExec)
	}
	// use config file if accessible, otherwise use a default conf
	// to mute zero exit codes
	conf, err := mute.GetCmdConf()
	if err != nil {
		if _, ok := err.(mute.ConfAccessError); ok {
			conf = mute.DefaultConf()
		} else {
			fmt.Fprintf(os.Stderr, "config error:  %v", err)
			os.Exit(mute.ExitErrConf)
		}
	}
	exitCode, _ := mute.Exec(os.Args[1], os.Args[2:], conf, os.Stdout, os.Stderr)
	os.Exit(exitCode)
}
