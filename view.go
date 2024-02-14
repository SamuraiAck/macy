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
	"code.rocketnine.space/tslocum/cview"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"sort"
	"time"
)

var (
	Log     = cview.NewTextView()
	Reports = cview.NewTable()
)

func View() {
	Log.SetBorder(true)
	Log.SetBorderColor(tcell.ColorGrey)
	Log.ShowFocus(false)
	Log.SetScrollBarColor(tcell.ColorGrey)
	Log.ScrollToEnd()
	Log.SetDynamicColors(true)

	Reports.SetBorder(true)
	Reports.SetBorderColor(tcell.ColorGrey)
	Reports.ShowFocus(false)
	Reports.SetScrollBarColor(tcell.ColorGrey)
	Reports.SetFixed(1, 1)
	Reports.SetSeparator(cview.Borders.Vertical)
	Reports.SetBordersColor(tcell.ColorGrey)

	about := cview.NewTextView()
	about.SetBorder(true)
	about.SetBorderColor(tcell.ColorGrey)
	about.ShowFocus(false)
	about.SetScrollBarColor(tcell.ColorGrey)
	about.SetDynamicColors(true)
	fmt.Fprint(about, License)

	quit := cview.NewBox()

	panels := cview.NewTabbedPanels()
	panels.SetTabSwitcherDivider("", "", "")
	panels.SetTabTextColor(tcell.ColorLightGrey)
	panels.SetTabBackgroundColor(tcell.ColorBlack)
	panels.SetTabTextColorFocused(tcell.ColorWhite)
	panels.SetTabBackgroundColorFocused(tcell.ColorGrey)
	panels.AddTab("Reports", "(R)eports", Reports)
	panels.AddTab("Log", "(L)og", Log)
	panels.AddTab("About", "(A)bout", about)
	panels.AddTab("Quit", "(Q)uit", quit)
	panels.SetCurrentTab("Reports")

	app := cview.NewApplication()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'r', 'R':
			panels.SetCurrentTab("Reports")
		case 'l', 'L':
			panels.SetCurrentTab("Log")
		case 'a', 'A':
			panels.SetCurrentTab("About")
		case 'q', 'Q':
			app.Stop()
		}

		return event
	})

	UpdateReports()
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			UpdateReports()
			app.Draw()
		}
	}()

	app.SetRoot(panels, true)
	err := app.Run()
	if err != nil {
		Fatal("%v", err)
	}
}

func UpdateReports() {
	Reports.Clear()

	// Lock access to maps
	Mutex.Lock()

	// Gather local IPs
	localIPs := make(map[string]bool)
	for _, s := range Senders {
		localIPs[s.IP.String()] = true
	}

	// Gather all IPs
	allIPs := make(map[string]bool)
	for ip := range localIPs {
		allIPs[ip] = true
	}
	for ip := range HeardIPs {
		allIPs[ip] = true
	}
	for _, heard := range HeardDb {
		for ip := range heard {
			allIPs[ip] = true
		}
	}
	var ips []string
	for ip := range allIPs {
		ips = append(ips, ip)
	}
	sort.Strings(ips)

	// Gather hosts
	allHosts := make(map[string]bool)
	allHosts[Host] = true
	for host := range HeardHosts {
		allHosts[host] = true
	}
	var hosts []string
	for host := range allHosts {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	if len(hosts) != 0 && len(ips) != 0 {
		// Gather data
		now := time.Now()
		data := [][]string{append([]string{""}, hosts...)}
		for _, ip := range ips {
			row := []string{ip}
			for _, host := range hosts {
				var t time.Time
				var d time.Duration

				if host == Host {
					t = HeardIPs[ip]
					if !t.IsZero() {
						d = now.Sub(t)
					}
				} else {
					d = HeardDb[host][ip]
					if d != 0 {
						d += now.Sub(HeardHosts[host])
					}
				}

				if d == 0 {
					row = append(row, "")
				} else {
					row = append(row, fmt.Sprintf("%.3fs", d.Seconds()))
				}
			}
			data = append(data, row)
		}

		// Update table
		for r, row := range data {
			for c, s := range row {
				cell := cview.NewTableCell(s)
				cell.SetAlign(cview.AlignRight)
				if r == 0 && s == Host {
					cell.SetTextColor(tcell.ColorAqua)
				}
				if c == 0 {
					cell.SetAlign(cview.AlignLeft)
					if localIPs[s] {
						cell.SetTextColor(tcell.ColorAqua)
					}
				}
				Reports.SetCell(r, c, cell)
			}
		}
	}

	// Unlock access to maps
	Mutex.Unlock()
}
