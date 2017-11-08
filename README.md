# sd-step
[![Build Status][build-image]][build-url]
[![Latest Release][version-image]][version-url]
[![Go Report Card][goreport-image]][goreport-url]

> Wrapper command of habitat for Screwdriver

## Usage

```bash
$ go get github.com/screwdriver-cd/sd-step
$ cd $GOPATH/src/github.com/screwdriver-cd/sd-step
$ go build -a -o sd-step
$ ./sd-step --help
NAME:
   sd-step - wrapper command of habitat for Screwdriver

USAGE:
   sd-step command arguments [options]

VERSION:
   0.0.0

COMMANDS:
     exec     Install and exec habitat package with pkg_name and command...
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --pkg-version value  Package version which also accepts semver expression
   --hab-channel value  Install from the specified release channel (default: "stable")
   --help, -h           show help
   --version, -v        print the version

COPYRIGHT:
   (c) 2017 Yahoo Inc.
$ ./sd-step exec core/node "node -v"
v8.9.0
$ ./sd-step exec --pkg-version "~6.11.0" core/node "node -v"
v6.11.5
$ ./sd-step exec --pkg-version "^6.0.0" core/node "node -v"
v6.11.5
$ ./sd-step exec --pkg-version "4.2.6" core/node "node -v"
v4.2.6
$ ./sd-step exec --pkg-version "~6.9.0" --hab-channel "unstable" core/node "node -v"
v6.9.5
```

## Testing

```bash
$ go get github.com/screwdriver-cd/sd-step
$ go test -cover github.com/screwdriver-cd/sd-step/...
```

## License

Code licensed under the BSD 3-Clause license. See LICENSE file for terms.

[version-image]: https://img.shields.io/github/tag/screwdriver-cd/sd-step.svg
[version-url]: https://github.com/screwdriver-cd/sd-step/releases
[build-image]: https://cd.screwdriver.cd/pipelines/150/badge
[build-url]: https://cd.screwdriver.cd/pipelines/150
[goreport-image]: https://goreportcard.com/badge/github.com/Screwdriver-cd/sd-step
[goreport-url]: https://goreportcard.com/report/github.com/Screwdriver-cd/sd-step
