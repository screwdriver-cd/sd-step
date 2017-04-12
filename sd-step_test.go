package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

const habExecResult = "run hab pkg install\nrun hab pkg exec\n"

func TestRunCommand(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	stdout := new(bytes.Buffer)
	err := runCommand("hab pkg install foo/bar", stdout)
	expected := "run hab pkg install\n"
	if err != nil {
		t.Errorf("runCommand error = %q, should be nil", err)
	}
	if string(stdout.Bytes()) != expected {
		t.Errorf("Expected '%v', actual '%v'", expected, string(stdout.Bytes()))
	}

	stdout = new(bytes.Buffer)
	err = runCommand("hab pkg exec foo/bar foo bar foobar", stdout)
	expected = "run hab pkg exec\n"
	if err != nil {
		t.Errorf("runCommand error = %v, should be nil", err)
	}
	if string(stdout.Bytes()) != expected {
		t.Errorf("Expected '%v', actual '%v'", expected, string(stdout.Bytes()))
	}
}

func TestExecHab(t *testing.T) {
	stdout := new(bytes.Buffer)
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	err := execHab("foo/bar", "2.2.2", []string{"foo", "bar", "foobar"}, stdout)
	if err != nil {
		t.Errorf("execHab error = %q, should be nil", err)
	}
	if string(stdout.Bytes()) != habExecResult {
		t.Errorf("Expected %q, got %q", habExecResult, string(stdout.Bytes()))
	}
}

func TestMain(m *testing.M) {
	retCode := m.Run()
	os.Exit(retCode)
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)
	args := os.Args[:]
	for i, val := range os.Args {
		args = os.Args[i:]
		if val == "-c" {
			args = strings.Split(args[1:][0], " ")
			break
		}
	}
	if len(args) >= 4 && args[1] == "pkg" {
		switch args[2] {
		case "install":
			fmt.Println("run hab pkg install")
			return
		case "exec":
			fmt.Println("run hab pkg exec")
			return
		default:
			os.Exit(255)
		}
	}
	os.Exit(255)
}
