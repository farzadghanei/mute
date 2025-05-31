// Package mute executes programs suppressing std streams if required
// license: MIT, see LICENSE for details.
// BurntSushi/toml module uses MIT license. see LICENSE for more details
package mute

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// Version is the program version
const Version string = "0.3.0"

// EnvConfig is the name of the environment variable to point to config file
const EnvConfig string = "MUTE_CONFIG"

// EnvExitCodes is the name of the environment variable to overwrite exit codes
const EnvExitCodes string = "MUTE_EXIT_CODES"

// EnvStdoutPattern is the name of the environment variable to overwrite stdout regex pattern
const EnvStdoutPattern string = "MUTE_STDOUT_PATTERN"

// ExitErrConf is exit code when config is invalid
const ExitErrConf = 126

// StdoutPattern hold regex pattern to match stdout with
type StdoutPattern struct {
	Regexp *regexp.Regexp
}

// UnmarshalText reads the regex pattern from a byte slice
func (s *StdoutPattern) UnmarshalText(text []byte) error {
	var err error
	re, err := regexp.Compile(string(text))
	s.Regexp = re
	return err
}

// String return the regex pattern string
func (s *StdoutPattern) String() string {
	return s.Regexp.String()
}

// NewStdoutPattern returns a pointer to a StdoutPattern object using the regex pattern string
func NewStdoutPattern(pattern string) *StdoutPattern {
	var stdp StdoutPattern
	re := regexp.MustCompile(pattern)
	stdp = StdoutPattern{Regexp: re}
	return &stdp
}

// Criterion is expected exit codes and stdout patterns to mute a process
type Criterion struct {
	ExitCodes      []int            `toml:"exit_codes"`
	StdoutPatterns []*StdoutPattern `toml:"stdout_patterns"`
}

// Criterion.String return a string desc to help debugging and inspecting data
// This string is not guaranteed to be structured nor accurate
func (c *Criterion) String() string {
	var codes []string

	for _, code := range c.ExitCodes {
		codes = append(codes, strconv.Itoa(code))
	}

	return fmt.Sprintf("<Criterion codes=\"%s\" patterns_count=\"%d\">", strings.Join(codes, ","), len(c.StdoutPatterns))
}

// IsEmpty checks if a Criterion is empty (no exit codes, no patterns)
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

// ConfAccessError represents errors when accessing to Config files
type ConfAccessError struct {
	err  error
	Path string
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
	var c1Crt, c2Crt Criteria
	var cmd string
	var ok bool
	if !c.Default.equal(&(c2.Default)) {
		return false
	}
	if len(c.Commands) != len(c2.Commands) {
		return false
	}
	for cmd, c1Crt = range c.Commands {
		if c2Crt, ok = c2.Commands[cmd]; !ok {
			return false
		}
		if !c1Crt.equal(&c2Crt) {
			return false
		}
	}
	return true
}

// IsEmpty determines if the Conf is empty
func (c *Conf) IsEmpty() bool {
	return len(c.Default) < 1 && len(c.Commands) < 1
}

func (e ConfAccessError) Error() string {
	return e.err.Error()
}

// ReadConfFile reads config file and returns Conf
func ReadConfFile(path string) (*Conf, error) {
	var conf Conf
	content, err := os.ReadFile(path)
	if err != nil {
		return &conf, ConfAccessError{err: err, Path: path}
	}
	contentStr := string(content)
	_, err = toml.Decode(contentStr, &conf)
	return &conf, err
}

// ConfFromEnvStr returns a Conf populated by strings as accepted environment variables
// If the strings are empty, and empty Conf with no Criterion will be returned
func ConfFromEnvStr(exitCodesStr, pattern string) (*Conf, error) {
	var err error
	var exitCodes []int
	var stdp StdoutPattern
	var reg *regexp.Regexp
	criterion := new(Criterion)
	conf := new(Conf)

	if exitCodesStr != "" {
		codeStrList := strings.Split(exitCodesStr, ",")
		for _, s := range codeStrList {
			i, err := strconv.Atoi(s)
			if err != nil {
				return conf, err
			}
			exitCodes = append(exitCodes, i)
		}
	}
	criterion.ExitCodes = exitCodes

	if pattern != "" {
		if reg, err = regexp.Compile(pattern); err != nil {
			return conf, err
		}
		stdp = StdoutPattern{Regexp: reg}
		criterion.StdoutPatterns = []*StdoutPattern{&stdp}
	}

	if !criterion.IsEmpty() {
		conf.Default.add(criterion)
	}

	return conf, err
}

// GetCmdConf returns the Conf that the mute cmd will use based on env vars
func GetCmdConf() (*Conf, error) {
	var conf *Conf
	var err error

	envExitCodes := os.Getenv(EnvExitCodes)
	envPattern := os.Getenv(EnvStdoutPattern)
	conf, err = ConfFromEnvStr(envExitCodes, envPattern)

	if err != nil || !conf.IsEmpty() {
		return conf, err
	}

	confPath := "/etc/mute.toml"
	envConfPath, envConfSet := os.LookupEnv(EnvConfig)
	if envConfSet {
		confPath = envConfPath
		if confPath == "" {
			conf = DefaultConf()
			return conf, err
		}
	}

	conf, err = ReadConfFile(confPath)
	return conf, err
}
