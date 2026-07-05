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
	"io"
	"net/http"
	"strings"
)

type MyIpCommand struct{}

func (v MyIpCommand) Execute(args []string) error {
	if len(args) > 0 {
		return errors.New("MyIp command does not accept any arguments. Try 'prox myip help' for usage information.")
	}
	url := "https://icanhazip.com"

	resp, err := http.Get(url)
	if err != nil {
		return errors.New("Request failed: " + err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Request failed: " + err.Error())
	}

	publicIP := strings.TrimSpace(string(body))
	if isPiped() {
		PrintInfo("%s", publicIP)
	} else {
		PrintInfo("Your public IP address is: %s", publicIP)
	}
	return nil
}
func (v MyIpCommand) Description() string {
	return "Display public ip address"
}
func init() {
	register("myip", MyIpCommand{})
}
