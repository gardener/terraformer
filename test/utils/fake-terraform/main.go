// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	exitCode      string
	sleepDuration string
)

// This packages contains a simple program which can be built in tests to mock terraform executions
// It basically just writes some lines to stdout and stderr, sleeps for `sleepDuration` and exits with `exitCode`.
func main() {
	fmt.Println("some terraform output")
	fmt.Println("args: " + strings.Join(os.Args[1:], " "))

	code, err := strconv.Atoi(exitCode)
	if err != nil {
		panic(err)
	}

	if sleepDuration != "" {
		duration, err := time.ParseDuration(sleepDuration)
		if err != nil {
			panic(err)
		}

		fmt.Printf("doing some long running IaaS ops for %s\n", duration.String())
		time.Sleep(duration)
	}

	fmt.Println("finished terraform execution")
	_, _ = fmt.Fprintln(os.Stderr, "some terraform error")

	if exitCode == "" {
		os.Exit(0)
	}
	os.Exit(code)
}
