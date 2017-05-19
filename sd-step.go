package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/screwdriver-cd/sd-step/hab"
	"github.com/urfave/cli"
)

// VERSION gets set by the build script via the LDFLAGS
var VERSION string

// HAB_DEPOT_URL is base url for public depot of habitat
const habDepotURL = "https://willem.habitat.sh/v1/depot"

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

	installCmd := []string{habPath, "pkg", "install", pkg, ">/dev/null"}
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

// getPackageVersion returns the appropriate package version which matched the `pkgVerExp` expression.
func getPackageVersion(depot hab.Depot, pkgName, pkgVerExp string) (string, error) {
	versionConst, err := semver.NewConstraint(pkgVerExp)
	// if pkgVerExp is invalid for semver expression, it returns pkgVerExp as it is
	if err != nil {
		return pkgVerExp, nil
	}

	foundVersions, err := depot.PackageVersionsFromName(pkgName)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch package versions: %v", err)
	}

	var versions []*semver.Version
	for _, version := range foundVersions {
		// if version exactly matches pkgVersionExp, it returns the version
		if version == pkgVerExp {
			return version, nil
		}

		v, err := semver.NewVersion(version)
		if err != nil {
			continue
		}

		if versionConst.Check(v) {
			versions = append(versions, v)
		}
	}

	if len(versions) == 0 {
		return "", errors.New("The specified version not found")
	}

	sort.Sort(sort.Reverse(semver.Collection(versions)))

	return versions[0].String(), nil
}

func main() {
	defer finalRecover()

	var pkgVerExp string

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
			Usage:       "Package version which also accepts semver expression",
			Value:       "",
			Destination: &pkgVerExp,
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

				pkgName := c.Args().Get(0)

				depot := hab.New(habDepotURL)
				pkgVersion, err := getPackageVersion(depot, pkgName, pkgVerExp)

				if err != nil {
					failureExit(fmt.Errorf("Failed to get package version: %v", err))
				}

				err = execHab(pkgName, pkgVersion, c.Args().Tail(), os.Stdout)
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
