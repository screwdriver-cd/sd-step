package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/urfave/cli"
)

// VERSION gets set by the build script via the LDFLAGS
var VERSION string

var habPath = "/opt/sd/bin/hab"
var versionValidator = regexp.MustCompile(`^\d+(\.\d+)*$`)
var execCommand = exec.Command

// successExit exits process with 0
func successExit() {
	os.Exit(0)
}

// failureExit exits process with 1
func failureExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
	os.Exit(1)
}

// finalRecover makes one last attempt to recover from a panic.
// This should only happen if the previous recovery caused a panic.
func finalRecover() {
	if p := recover(); p != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Something terrible has happened. Please file a ticket with this info:")
		fmt.Fprintf(os.Stderr, "ERROR: %v\n%v\n", p, debug.Stack())
		failureExit(nil)
	}
	successExit()
}

// translatePkgName translates the pkgName if pkgVersion is specified
func translatePkgName(pkgName string, pkgVersion string) (string, error) {
	if pkgVersion == "" {
		return pkgName, nil
	} else if valid := versionValidator.MatchString(pkgVersion); valid == true {
		return pkgName + "/" + pkgVersion, nil
	} else {
		return "", fmt.Errorf("%v is invalid version", pkgVersion)
	}
}

// runCommand runs command
func runCommand(command string, output io.Writer) error {
	cmd := execCommand("sh", "-c", command)
	cmd.Stdout = output
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// execHab installs habitat package and executes habitat command
func execHab(pkgName string, pkgVersion string, command []string, output io.Writer) error {
	pkg, verErr := translatePkgName(pkgName, pkgVersion)
	if verErr != nil {
		return verErr
	}

	installCmd := []string{habPath, "pkg", "install", pkg}
	unwrappedInstallCommand := strings.Join(installCmd, " ")
	installErr := runCommand(unwrappedInstallCommand, output)
	if installErr != nil {
		return installErr
	}

	execCmd := []string{habPath, "pkg", "exec", pkg}
	unwrappedExecCommand := strings.Join(append(execCmd, command...), " ")
	execErr := runCommand(unwrappedExecCommand, output)
	if execErr != nil {
		return execErr
	}

	return nil
}

func main() {
	defer finalRecover()

	var pkgVersion string

	app := cli.NewApp()
	app.Name = "sd-step"
	app.Usage = "wrapper command of habitat for Screwdriver"
	app.UsageText = "sd-step command arguments [options]"
	app.Copyright = "(c) 2017 Yahoo Inc."

	if VERSION == "" {
		VERSION = "0.0.0"
	}
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "pkg-version",
			Usage:       "Package version",
			Value:       "",
			Destination: &pkgVersion,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "exec",
			Usage: "Install and exec habitat package with pkg_name and command...",
			Action: func(c *cli.Context) error {
				if len(c.Args()) < 2 {
					return cli.ShowAppHelp(c)
				}
				err := execHab(c.Args().Get(0), pkgVersion, c.Args().Tail(), os.Stdout)
				if err != nil {
					failureExit(err)
				}
				successExit()
				return nil
			},
			Flags: app.Flags,
		},
	}

	app.Run(os.Args)
}
