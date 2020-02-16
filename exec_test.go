// license: MIT, see LICENSE for details.
package mute

import (
	"bufio"
	"bytes"
	"testing"
)

func TestExecMute(t *testing.T) {
	conf := DefaultConf()
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	outWriter := bufio.NewWriter(&outBuf)
	errWriter := bufio.NewWriter(&errBuf)
	code, err := Exec("go", []string{"version"}, conf, outWriter, errWriter)
	outWriter.Flush()
	errWriter.Flush()
	if err != nil {
		t.Errorf("Exec returned error")
	}
	if code != 0 {
		t.Errorf("Exec return val. got: %d want: 0", code)
	}
	outStr := outBuf.String()
	errStr := errBuf.String()
	if outStr != "" {
		t.Errorf("Exec mute should not print output. got: %v", outStr)
	}
	if errStr != "" {
		t.Errorf("Exec mute should not print error. got: %v", errStr)
	}
}

func TestExecNoMute(t *testing.T) {
	conf := DefaultConf()
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	outWriter := bufio.NewWriter(&outBuf)
	errWriter := bufio.NewWriter(&errBuf)
	code, err := Exec("go", []string{"invalid"}, conf, outWriter, errWriter)
	outWriter.Flush()
	errWriter.Flush()
	if err == nil {
		t.Errorf("Exec invalid didn't return error")
	}
	if code == 0 {
		t.Errorf("Exec invalid return val is 0")
	}
	errStr := errBuf.String()
	if errStr == "" {
		t.Errorf("Exec invalid should print error but didn't")
	}
}

func TestCmdCriteriaReturnDefault(t *testing.T) {
	c1 := NewCriterion([]int{0}, []string{})
	conf := new(Conf)
	conf.Default.add(c1)

	got := cmdCriteria("testcommand", conf)
	if !got.equal(&conf.Default) {
		t.Errorf("cmdCriteria should have returned conf default but didn't")
	}
}

func TestCmdCriteriaReturnCommandSpecific(t *testing.T) {
	var crt1, crt2 Criteria
	c1 := NewCriterion([]int{0}, []string{})
	c2 := NewCriterion([]int{1}, []string{})
	crt1 = append(crt1, c1)
	crt2 = append(crt2, c2)

	var commandsCriteria map[string]Criteria
	commandsCriteria = make(map[string]Criteria)
	commandsCriteria["test"] = crt1
	commandsCriteria["testcommand"] = crt2
	commandsCriteria["somethingelse"] = crt1

	conf := new(Conf)
	conf.Commands = commandsCriteria

	got := cmdCriteria("testcommand", conf)
	if !got.equal(&crt2) {
		t.Errorf("cmdCriteria should have returned longest matched cmd but didn't")
	}
}

func TestMatchesCriteria(t *testing.T) {
	conf, _ := ReadConfFile("fixtures/simple.toml")
	crt := conf.Default
	stdout := ""
	if !matchesCriteria(&crt, 0, &stdout) {
		t.Errorf("matchesCriteria 0 default want 'true' got 'false'")
	}
	if matchesCriteria(&crt, 3, &stdout) {
		t.Errorf("matchesCriteria 3 default want 'false' got 'true'")
	}
	if matchesCriteria(&crt, 1, &stdout) {
		t.Errorf("matchesCriteria 1 empty stdout want 'false' got 'true'")
	}
	stdout = "OK"
	if !matchesCriteria(&crt, 1, &stdout) {
		t.Errorf("matchesCriteria 1 matching stdout want 'true' got 'false'")
	}
}
