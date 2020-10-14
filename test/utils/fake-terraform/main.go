// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
