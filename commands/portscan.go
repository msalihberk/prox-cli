/* Copyright 2026 Mustafa Salih Berk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package commands

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PortscanCommand struct{}

func parsePorts(portStr string) ([]int, error) {
	var ports []int
	if portStr == "" {
		commonPorts := []int{21, 22, 23, 25, 53, 80, 110, 135, 139, 143, 443, 445, 993, 995, 1723, 3306, 3389, 5900, 8080, 8443}
		return commonPorts, nil
	}

	parts := strings.Split(portStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid port range: %s", part)
			}
			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, err
			}
			for i := start; i <= end; i++ {
				ports = append(ports, i)
			}
		} else {
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			ports = append(ports, port)
		}
	}
	return ports, nil
}

func (v PortscanCommand) Execute(args []string) error {
	parser := New(args, false)
	parser.Parse()

	if len(args) == 0 {
		return errors.New("portscan command requires a target host. Try 'prox portscan help' for usage information.")
	}

	target, _ := parser.Pos(0)
	if target == "help" || parser.GetAlias("h", "help").Found {
		PrintInfo("%s", v.Help())
		return nil
	}

	portParam := ""
	if parser.GetAlias("p", "ports").Found {
		portParam = parser.GetAlias("p", "ports").Value
	}

	portsToScan, err := parsePorts(portParam)
	if err != nil {
		return errors.New("failed to parse ports: " + err.Error())
	}

	workerCount := 100
	if parser.GetAlias("w", "workers").Found {
		if val, err := strconv.Atoi(parser.GetAlias("w", "workers").Value); err == nil && val > 0 {
			workerCount = val
		}
	}

	timeoutMs := 500
	if parser.GetAlias("t", "timeout").Found {
		if val, err := strconv.Atoi(parser.GetAlias("t", "timeout").Value); err == nil && val > 0 {
			timeoutMs = val
		}
	}
	timeout := time.Duration(timeoutMs) * time.Millisecond

	if !isPiped() {
		PrintMessage("Scanning %s (%d ports) using %d workers...", target, len(portsToScan), workerCount)
	}

	portsChan := make(chan int, workerCount)
	resultsChan := make(chan int)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range portsChan {
				address := net.JoinHostPort(target, strconv.Itoa(port))
				conn, err := net.DialTimeout("tcp", address, timeout)
				if err != nil {
					continue
				}
				conn.Close()
				resultsChan <- port
			}
		}()
	}

	var openPorts []int
	var resultsWg sync.WaitGroup
	resultsWg.Add(1)
	go func() {
		defer resultsWg.Done()
		for port := range resultsChan {
			openPorts = append(openPorts, port)
			if isPiped() {
				fmt.Printf("%d\n", port)
			}
		}
	}()

	for _, port := range portsToScan {
		portsChan <- port
	}
	close(portsChan)
	wg.Wait()

	close(resultsChan)
	resultsWg.Wait()

	if !isPiped() {
		sort.Ints(openPorts)
		PrintNewLine()
		if len(openPorts) == 0 {
			PrintWarning("No open ports found on %s", target)
		} else {
			PrintSuccess("Scan complete. Found %d open ports:", len(openPorts))
			for _, port := range openPorts {
				PrintInfo("  Port %d is OPEN", port)
			}
		}
	}

	return nil
}

func (v PortscanCommand) Description() string {
	return "Scan a target host for open ports concurrently"
}
func (v PortscanCommand) Help() string {
	help := "Usage: prox portscan <target> [-p <ports>] [-w <workers>] [-t <timeout>]"
	help += "\n  <target>          : Domain or IP address to scan"
	help += "\n  -p, --ports       : Ports to scan (e.g., '80,443' or '20-80'). Default: 20 common ports"
	help += "\n  -w, --workers     : Number of concurrent workers. Default: 100"
	help += "\n  -t, --timeout     : Timeout in milliseconds. Default: 500"
	return help
}
func init() {
	register("portscan", PortscanCommand{})
}
