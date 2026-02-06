package core

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestExitCode(t *testing.T) {
	ran, err := Exec(nil, nil, nil, "", "sh", "-c", "exit 99")
	if err == nil {
		t.Fatal("unexpected nil error from run")
	}
	if !ran {
		t.Errorf("ran returned as false, but should have been true")
	}
	code := ExitStatus(err)
	if code != 99 {
		t.Fatalf("expected exit status 99, but got %v", code)
	}
}

func TestSettingPwd(t *testing.T) {
	pwd := "/"
	out := &bytes.Buffer{}
	ran, err := Exec(nil, out, nil, pwd, "pwd")
	if err != nil {
		t.Fatalf("unexpected error from runner: %#v", err)
	}
	if !ran {
		t.Errorf("expected ran to be true but was false.")
	}
	if out.String() != fmt.Sprintf("%s\n", pwd) {
		t.Errorf("expected %s, got %q", fmt.Sprintf("%s\n", pwd), out)
	}
}

func TestSettingNoPwd(t *testing.T) {
	currentWd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed getting current working directory")
	}
	out := &bytes.Buffer{}
	ran, err := Exec(nil, out, nil, "", "pwd")
	if err != nil {
		t.Fatalf("unexpected error from runner: %#v", err)
	}
	if !ran {
		t.Errorf("expected ran to be true but was false.")
	}
	if out.String() != fmt.Sprintf("%s\n", currentWd) {
		t.Errorf("expected %s, got %q", fmt.Sprintf("%s\n", currentWd), out)
	}
}

func TestSettingInvalidPwd(t *testing.T) {
	pwd := "/i-am-expected-to-not-exist"
	out := &bytes.Buffer{}
	_, err := Exec(nil, out, nil, pwd, "pwd")
	if err == nil {
		t.Fatalf("I am expected to fail because path %s does not exist", pwd)
	}
}

func TestEnv(t *testing.T) {
	env := "SOME_REALLY_LONG_MAGEFILE_SPECIFIC_THING"
	out := &bytes.Buffer{}
	ran, err := Exec(map[string]string{env: "foobar"}, out, nil, "", "echo", fmt.Sprintf("$%s", env))
	if err != nil {
		t.Fatalf("unexpected error from runner: %#v", err)
	}
	if !ran {
		t.Errorf("expected ran to be true but was false.")
	}
	if out.String() != "foobar\n" {
		t.Errorf("expected foobar, got %q", out)
	}
}

func TestNotRun(t *testing.T) {
	ran, err := Exec(nil, nil, nil, "", "thiswontwork")
	if err == nil {
		t.Fatal("unexpected nil error")
	}
	if ran {
		t.Fatal("expected ran to be false but was true")
	}
}

func TestAutoExpand(t *testing.T) {
	if err := os.Setenv("MAGE_FOOBAR", "baz"); err != nil {
		t.Fatal(err)
	}
	s, err := Output("echo", "$MAGE_FOOBAR")
	if err != nil {
		t.Fatal(err)
	}
	if s != "baz" {
		t.Fatalf(`Expected "baz" but got %q`, s)
	}
}
