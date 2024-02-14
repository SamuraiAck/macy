// Copyright 2024 Eric Johnson
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"github.com/muesli/termenv"
	"os"
	"strings"
)

var (
	Colors        = termenv.ColorProfile()
	FatalLog      strings.Builder
	LogCandidates = make(map[string]string)
)

func Aqua(s string) string {
	return termenv.String(s).Foreground(Colors.Color("#2aa1b3")).String()
}
func Green(s string) string {
	return termenv.String(s).Foreground(Colors.Color("#26a269")).String()
}
func Yellow(s string) string {
	return termenv.String(s).Foreground(Colors.Color("#c4a90e")).String()
}
func Red(s string) string {
	return termenv.String(s).Foreground(Colors.Color("#f15f42")).String()
}

func Debug(format string, args ...any) {
	if Verbose {
		s := fmt.Sprintf(format, args...)
		fmt.Fprintf(Log, "[aqua]DEBUG[white] %s\n", s)
		FatalLog.WriteString(fmt.Sprintf("%s %s\n", Aqua("DEBUG"), s))
	}
}

func Info(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	fmt.Fprintf(Log, "[green]INFO[white] %s\n", s)
	FatalLog.WriteString(fmt.Sprintf("%s %s\n", Green("INFO"), s))
}

func Warn(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	fmt.Fprintf(Log, "[yellow]WARN[white] %s\n", s)
	FatalLog.WriteString(fmt.Sprintf("%s %s\n", Yellow("WARN"), s))
}

func Fatal(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	FatalLog.WriteString(fmt.Sprintf("%s %s\n", Red("FATAL"), s))
	fmt.Print(FatalLog.String())
	os.Exit(0)
}
