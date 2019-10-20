package mute

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

// Criterion is expected exit codes and stdout patterns to mute a process
type Criterion struct {
	ExitCodes      []int    `toml:"exit_codes"`
	StdoutPatterns []string `toml:"stdout_patterns"`
}

// Criteria is a list of Criterion that if a process matched any of, it'll be muted
type Criteria []*Criterion

// Conf is the mute configuration of default and per process criteria
type Conf struct {
	Default  Criteria
	Commands map[string]Criteria
}

func NewCriterion(codes []int, patterns []string) *Criterion {
	c := new(Criterion)
	c.ExitCodes = codes
	c.StdoutPatterns = patterns
	return c
}

func (c *Criteria) add(items ...*Criterion) *Criteria {
	*c = append(*c, items...)
	return c
}

// codesContain search for a given exit code in a slice of exit codes
func codesContain(codes []int, code int) bool {
	for _, item := range codes {
		if item == code {
			return true
		}
	}
	return false
}

// stringsContains search through a slice os strings for a given string
func stringsContain(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

// Criterion.equal returns true of Criterions have the same exit codes and stdout patterns
func (c *Criterion) equal(c2 *Criterion) bool {
	if len(c.ExitCodes) != len(c2.ExitCodes) || len(c.StdoutPatterns) != len(c2.StdoutPatterns) {
		return false
	}
	for _, code := range c.ExitCodes {
		if !codesContain(c2.ExitCodes, code) {
			return false
		}
	}
	for _, pattern := range c.StdoutPatterns {
		if !stringsContain(c2.StdoutPatterns, pattern) {
			return false
		}
	}
	return true
}

// Criteria.contains check if the criteria contains a given criterion
func (c *Criteria) contains(criterion *Criterion) bool {
	for _, item := range *c {
		if criterion.equal(item) {
			return true
		}
	}
	return false
}

// Criteria.equal check if criteria items are the same
func (c *Criteria) equal(c2 *Criteria) bool {
	if len(*c) != len(*c2) {
		return false
	}
	for _, item := range *c {
		if !c2.contains(item) {
			return false
		}
	}
	return true
}

// Conf.equal check if Conf items are the same
func (c *Conf) equal(c2 *Conf) bool {
	return c.Default.equal(&(c2.Default))
}

// ReadConfFile reads config file and returns Conf
func ReadConfFile(path string) (Conf, error) {
	var conf Conf
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return conf, err
	}
	contentStr := string(content)
	_, err = toml.Decode(contentStr, &conf)
	return conf, err
}
