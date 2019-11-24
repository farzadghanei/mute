package mute

import (
	"os"
	"testing"
)

func TestCodesContain(t *testing.T) {
	haystack := []int{1, 2}
	if !codesContain(haystack, 1) {
		t.Errorf("codesContain got 'false' want 'true'")
	}
	if codesContain(haystack, 3) {
		t.Errorf("codesContain got 'true' want 'false'")
	}
}

func TestStdoutPatternsContain(t *testing.T) {
	stdp1 := NewStdoutPattern("hi")
	stdp2 := NewStdoutPattern(".+not[1-9]+so.*simple")
	stdp3 := NewStdoutPattern("I was 3rd")
	stdp4 := NewStdoutPattern(".+not[1-9]+so.*close")

	haystack := []*StdoutPattern{stdp1, stdp2, stdp3}
	if !stdoutPatternsContain(haystack, stdp2) {
		t.Errorf("stdoutPatternsContain got 'false' want 'true'")
	}
	if stdoutPatternsContain(haystack, stdp4) {
		t.Errorf("stdoutPatternsContain got 'true' want 'false'")
	}
}

func TestCriterionEmpty(t *testing.T) {
	c1 := NewCriterion([]int{0}, []string{})
	c2 := NewCriterion([]int{}, []string{""})
	c3 := NewCriterion([]int{}, []string{})

	if c1.IsEmpty() {
		t.Errorf("Criterion with exit codes IsEmpty got 'true' want 'false'")
	}
	if c2.IsEmpty() {
		t.Errorf("Criterion with patterns IsEmpty got 'true' want 'false'")
	}
	if !c3.IsEmpty() {
		t.Errorf("Empty Criterion.IsEmpty 'false' want 'true'")
	}
}

func TestCriterionEqual(t *testing.T) {
	c1 := NewCriterion([]int{0, 1, 2}, []string{})
	c2 := NewCriterion([]int{0, 1, 2}, []string{})

	if !c1.equal(c2) {
		t.Errorf("Criterion.equal got 'false' want 'true'")
	}

	c2 = NewCriterion([]int{0, 2}, []string{})
	if c1.equal(c2) {
		t.Errorf("Criterion.equal unmatched codes got 'true' want 'false'")
	}

	c2 = NewCriterion([]int{0, 1, 2}, []string{"ok"})
	if c1.equal(c2) {
		t.Errorf("Criterion.equal unmatched patterns got 'true' want 'false'")
	}
}

func TestCriterionString(t *testing.T) {
	var want, got string
	var c1 *Criterion

	c1 = NewCriterion([]int{0}, []string{})
	want = "<Criterion codes=\"0\" patterns_count=\"0\">"
	got = c1.String()

	if want != got {
		t.Errorf("Criterion String failed. want: '%v' got '%v'", want, got)
	}

	c1 = NewCriterion([]int{0}, []string{"OK"})
	want = "<Criterion codes=\"0\" patterns_count=\"1\">"
	got = c1.String()

	if want != got {
		t.Errorf("Criterion String failed. want: '%v' got '%v'", want, got)
	}
}

func TestCriteriaEqual(t *testing.T) {
	c1 := NewCriterion([]int{0, 1, 2}, []string{})
	c2 := NewCriterion([]int{0, 1, 2}, []string{})

	criteria1 := new(Criteria)
	criteria1.add(c1, c2)

	criteria2 := new(Criteria)
	criteria2.add(c1, c2)

	if !c1.equal(c2) {
		t.Errorf("Criteria.equal got 'false' want 'true'")
	}

	c3 := NewCriterion([]int{0, 1, 2}, []string{})
	criteria2.add(c3)

	if !c1.equal(c2) {
		t.Errorf("Criteria.equal got 'true' want 'false'")
	}
}

func TestDefaultConf(t *testing.T) {
	c1 := NewCriterion([]int{0}, []string{})
	crt1 := new(Criteria)
	crt1.add(c1)

	conf := DefaultConf()
	if !conf.Default.equal(crt1) {
		t.Errorf("DefaultConf().Default didn't match zero exit code")
	}
}

func TestConfEqual(t *testing.T) {
	crt := NewCriterion([]int{1, 2}, []string{"OK"})

	empty1 := new(Conf)
	empty2 := new(Conf)
	simple1 := createSimpleConf()
	simple2 := createSimpleConf()

	cmd1 := createSimpleConf()
	cmd1.Commands["/usr/local/bin/mute"] = Criteria{crt}

	cmd2 := createSimpleConf()
	cmd2.Commands["/usr/local/bin/mute"] = Criteria{crt}

	if !empty1.equal(empty2) {
		t.Errorf("Conf empty1 should be equal to empty2")
	}

	if !simple1.equal(simple2) {
		t.Errorf("Conf simple should be equal to simple2")
	}

	if empty1.equal(simple1) {
		t.Errorf("Conf empty should not be equal to simple")
	}

	if !cmd1.equal(cmd2) {
		t.Errorf("Conf cmd1 should be equal to cmd2")
	}

	cmd2.Commands["test"] = Criteria{crt}
	if cmd1.equal(cmd2) {
		t.Errorf("Conf cmd1 should not be equal to cmd2 with extra command")
	}
}

func TestReadConfFileError(t *testing.T) {
	_, err := ReadConfFile("fixtures/no_such_file.toml")
	if err == nil {
		t.Errorf("ReadConfFile should have returned error")
	}
	if _, ok := err.(ConfAccessError); !ok {
		t.Errorf("ReadConfFile should have returned ConfAccessError")
	}
}

func TestReadConfFileSimple(t *testing.T) {
	c1 := NewCriterion([]int{0}, []string{})
	c2 := NewCriterion([]int{1, 2}, []string{"OK"})
	want := new(Conf)
	want.Default.add(c1).add(c2)

	got, err := ReadConfFile("fixtures/simple.toml")
	if err != nil {
		t.Errorf("ReadConfFile had error: %v", err)
	}

	if !want.equal(got) {
		t.Errorf("ReadConfFile simple didn't match want %v got %v", want, got)
	}

	c3 := NewCriterion([]int{1}, []string{})

	extraCodesConf := new(Conf)
	extraCodesConf.Default.add(c1, c3)

	if extraCodesConf.equal(got) {
		t.Errorf("ReadConfFile simple matched extra codes conf")
	}

	missingCodesConf := new(Conf)
	missingCodesConf.Default.add(c1)

	if missingCodesConf.equal(got) {
		t.Errorf("ReadConfFile simple matched missing codes conf")
	}
}

func TestConfFromEnvStr(t *testing.T) {
	var got, want, defaultConf *Conf
	var err error
	defaultConf = DefaultConf()

	got, err = ConfFromEnvStr("", "")
	if err != nil {
		t.Errorf("ConfFromEnvStr empty want no error, got: %v", err)
	}
	if !got.IsEmpty() {
		t.Errorf("ConfFromEnvStr want empty conf, got: %v", got)
	}

	got, err = ConfFromEnvStr("0", "")
	if err != nil {
		t.Errorf("ConfFromEnvStr default want no error, got: %v", err)
	}
	if !defaultConf.equal(got) {
		t.Errorf("ConfFromEnvStr want default conf, got: %v", got)
	}

	want = new(Conf)
	c1 := NewCriterion([]int{1, 2}, []string{"[0-9]test"})
	want.Default.add(c1)
	got, err = ConfFromEnvStr("1,2", "[0-9]test")
	if err != nil {
		t.Errorf("ConfFromEnvStr test want no error, got: %v", err)
	}
	if !want.equal(got) {
		t.Errorf("ConfFromEnvStr test want: %v, got: %v", want, got)
	}
}

func TestGetCmdConfFromEnv(t *testing.T) {
	var got, want *Conf
	var err error
	os.Setenv(ENV_CONFIG, "fixtures/simple.toml")
	defer os.Unsetenv(ENV_CONFIG)

	os.Unsetenv(ENV_EXIT_CODES)
	os.Unsetenv(ENV_STDOUT_PATTERN)

	want = createSimpleConf()
	got, err = GetCmdConf()
	if err != nil {
		t.Errorf("GetCmdConf no env conf want no error, got: %v", err)
	}
	if !want.equal(got) {
		t.Errorf("GetCmdConf no env conf want simple %v got %v", want, got)
	}

	os.Setenv(ENV_EXIT_CODES, "")
	defer os.Unsetenv(ENV_EXIT_CODES)
	os.Setenv(ENV_STDOUT_PATTERN, "")
	defer os.Unsetenv(ENV_STDOUT_PATTERN)

	got, err = GetCmdConf()
	if err != nil {
		t.Errorf("GetCmdConf empty env conf want no error, got: %v", err)
	}
	if !want.equal(got) {
		t.Errorf("GetCmdConf empty env conf want simple %v got %v", want, got)
	}

	c1 := NewCriterion([]int{4, 5}, []string{"[0-9]test"})
	want = new(Conf)
	want.Default.add(c1)

	os.Setenv(ENV_EXIT_CODES, "4,5")
	os.Setenv(ENV_STDOUT_PATTERN, "[0-9]test")
	got, err = GetCmdConf()

	if err != nil {
		t.Errorf("GetCmdConf env conf want no error, got: %v", err)
	}
	if !want.equal(got) {
		t.Errorf("GetCmdConf env want %v got %v", want, got)
	}

	os.Setenv(ENV_EXIT_CODES, "4,z")
	got, err = GetCmdConf()

	if err == nil {
		t.Errorf("GetCmdConf env inavlid exit code want error, got: %v", got)
	}
	if !got.IsEmpty() {
		t.Errorf("GetCmdConf env invalid exit code want empty conf, got %v", got)
	}

	os.Setenv(ENV_EXIT_CODES, "")
	os.Setenv(ENV_STDOUT_PATTERN, "[")
	got, err = GetCmdConf()

	if err == nil {
		t.Errorf("GetCmdConf env inavlid stdout pattern want error, got: %v", got)
	}
	if !got.IsEmpty() {
		t.Errorf("GetCmdConf env invalid stdout pattern want empty conf, got %v", got)
	}
}

func TestGetCmdConfFromFile(t *testing.T) {
	os.Setenv(ENV_CONFIG, "fixtures/simple.toml")
	defer os.Unsetenv(ENV_CONFIG)
	want := createSimpleConf()
	defaultConf := DefaultConf()
	got, err := GetCmdConf()
	if err != nil {
		t.Errorf("GetCmdConf simple want no error, got: %v", err)
	}
	if !want.equal(got) {
		t.Errorf("GetCmdConf simple conf want simple %v got %v", want, got)
	}

	os.Setenv(ENV_CONFIG, "")
	got, err = GetCmdConf()
	if err != nil {
		t.Errorf("GetCmdConf empty env want no error, got: %v", err)
	}
	if !defaultConf.equal(got) {
		t.Errorf("GetCmdConf empty env conf want default conf, got %v", got)
	}
}

// createSimpleConf returns a Conf with simple criterions for testing
func createSimpleConf() *Conf {
	c1 := NewCriterion([]int{0}, []string{})
	c2 := NewCriterion([]int{1, 2}, []string{"OK"})
	conf := new(Conf)
	conf.Default.add(c1).add(c2)
	conf.Commands = make(map[string]Criteria)
	return conf
}
