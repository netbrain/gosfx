# gosfx

A small utility to create self extracting archives with a entry point.

## Installation

`go get -u github.com/netbrain/gosfx/...`

## Example usage

Invoking `gosfx-packer -output ./my-application  -main "./entrypoint.sh" ./` would pack everything in the current folder inside the `my-application` executable.

When invoking the executable `my-application` the application launches and starts unpacking the files read from the binary to a temporary folder and with this folder as cwd invokes the `entrypoint.sh` command.

[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/netbrain/gosfx)
[![Go Report Card](https://goreportcard.com/badge/github.com/golang-standards/project-layout?style=flat-square)](https://goreportcard.com/report/github.com/netbrain/gosfx)
