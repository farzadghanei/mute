package mute

import (
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

func TestStringsContain(t *testing.T) {
	haystack := []string{"hi", " with spaces ", "?"}
	if !stringsContain(haystack, " with spaces ") {
		t.Errorf("stringsContain got 'false' want 'true'")
	}
	if stringsContain(haystack, "nope") {
		t.Errorf("stringsContain got 'true' want 'false'")
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

func TestReadConfFileError(t *testing.T) {
	_, err := ReadConfFile("fixtures/no_such_file.toml")
	if err == nil {
		t.Errorf("ReadConfFile should have returned error")
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

	if !want.equal(&got) {
		t.Errorf("ReadConfFileSimple didn't match want %v got %v", want, got)
	}

	c3 := NewCriterion([]int{1}, []string{})

	extraCodesConf := new(Conf)
	extraCodesConf.Default.add(c1, c3)

	if extraCodesConf.equal(&got) {
		t.Errorf("ReadConfFileSimple matched extra codes conf")
	}

	missingCodesConf := new(Conf)
	missingCodesConf.Default.add(c1)

	if missingCodesConf.equal(&got) {
		t.Errorf("ReadConfFileSimple matched missing codes conf")
	}
}
