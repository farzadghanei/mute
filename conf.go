package mute

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"regexp"
)

const Version string = "0.1.0"

// StdoutPattern hold regex pattern to match stdout with
type StdoutPattern struct {
	Regexp *regexp.Regexp
}

// StdoutPattern.UnmarshalText reads the regex pattern from a byte slice
func (s *StdoutPattern) UnmarshalText(text []byte) error {
	var err error
	re, err := regexp.Compile(string(text))
	s.Regexp = re
	return err
}

// NewStdoutPattern returns a pointer to a StdoutPattern object using the regex pattern string
func NewStdoutPattern(pattern string) *StdoutPattern {
	var stdp StdoutPattern
	var re *regexp.Regexp
	re = regexp.MustCompile(pattern)
	stdp = StdoutPattern{Regexp: re}
	return &stdp
}

// Criterion is expected exit codes and stdout patterns to mute a process
type Criterion struct {
	ExitCodes      []int            `toml:"exit_codes"`
	StdoutPatterns []*StdoutPattern `toml:"stdout_patterns"`
}

// Criterion.IsEmpty checks if a Criterion is empty (no exit codes, no patterns)
func (c *Criterion) IsEmpty() bool {
	return len(c.ExitCodes) < 1 && len(c.StdoutPatterns) < 1
}

// Criteria is a list of Criterion that if a process matched any of, it'll be muted
type Criteria []*Criterion

// Conf is the mute configuration of default and per process criteria
type Conf struct {
	Default  Criteria
	Commands map[string]Criteria
}

// NewCriterion returns pointer to Criterion with specified exit codes and regex patterns from strings
func NewCriterion(codes []int, patterns []string) *Criterion {
	c := new(Criterion)
	c.ExitCodes = codes
	var stdp *StdoutPattern
	for _, p := range patterns {
		stdp = NewStdoutPattern(p)
		c.StdoutPatterns = append(c.StdoutPatterns, stdp)
	}
	return c
}

// DefaultConf returns a Conf to use when there is no conf file to read
// It's a Conf with a Default Criteria to mute only successful
// runs (zero exit code)
func DefaultConf() *Conf {
	criterion := NewCriterion([]int{0}, []string{})
	conf := new(Conf)
	crt := &conf.Default
	crt.add(criterion)
	return conf
}

// Criteria.add adds more Criterions to current Criteria
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

// stdoutPatternsContain searches through a slice of regex patterns for a given regex patterns
func stdoutPatternsContain(haystack []*StdoutPattern, needle *StdoutPattern) bool {
	needleStr := needle.Regexp.String()
	for _, item := range haystack {
		if item.Regexp.String() == needleStr {
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
		if !stdoutPatternsContain(c2.StdoutPatterns, pattern) {
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
