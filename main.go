package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const baseURL = "https://api.ssllabs.com/api/v2"

type HostResponse struct {
	Host      string     `json:"host"`
	Status    string     `json:"status"`
	StatusMsg string     `json:"statusMessage"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	IPAddress     string           `json:"ipAddress"`
	StatusMessage string           `json:"statusMessage"`
	Grade         string           `json:"grade"`
	HasWarnings   bool             `json:"hasWarnings"`
	Details       *EndpointDetails `json:"details"`
}
type EndpointDetails struct {
	Protocols      []Protocol `json:"protocols"`
	ForwardSecrecy int        `json:"forwardSecrecy"`
	Heartbleed     bool       `json:"heartbleed"`
	Poodle         bool       `json:"poodle"`
	Logjam         bool       `json:"logjam"`
	Cert           Cert       `json:"cert"`
}

type Protocol struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Cert struct {
	CommonNames []string `json:"commonNames"`
	AltNames    []string `json:"altNames"`
	NotAfter    int64    `json:"notAfter"`
}

func analyze(host string, startNew bool) (*HostResponse, error) {
	url := fmt.Sprintf("%s/analyze?host=%s&all=done", baseURL, host)
	if startNew {
		url += "&startNew=on"
	}

	resp, err := http.Get(url)
	if err != nil {

		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result HostResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func readHostFromUser() string {

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("If you don't know a host use www.ssllabs.com")
		fmt.Print("Enter host: ")
		scanner.Scan()
		host := scanner.Text()

		if err := validateHost(host); err != nil {
			fmt.Println("Invalid host:", err)
			fmt.Println("Please try again.\n")
			continue
		}

		return host
	}
}

func validateHost(host string) error {
	host = strings.TrimSpace(host)

	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if strings.Contains(host, "://") {
		return fmt.Errorf("host must not include protocol (http/https)")
	}

	if strings.Contains(host, "/") {
		return fmt.Errorf("host must not contain path")
	}

	// Checks if format of host is valid
	if !isValidHostname(host) {
		return fmt.Errorf("invalid hostname format")
	}

	// checks if the host is in the DNS
	if _, err := net.LookupHost(host); err != nil {
		return fmt.Errorf("host does not resolve via DNS")
	}

	return nil
}

func isValidHostname(host string) bool {
	if len(host) > 253 {
		return false
	}

	labels := strings.Split(host, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, r := range label {
			if !(r >= 'a' && r <= 'z' ||
				r >= 'A' && r <= 'Z' ||
				r >= '0' && r <= '9' ||
				r == '-') {
				return false
			}
		}
	}
	return true
}

func yesNo(ok bool) string {
	if ok {
		return "OK"
	}
	return "VULNERABLE"
}

func main() {
	// Run test.go with www.ssllabs.com
	//fmt.Println("Starting test")
	//err := test.RunAllStates("www.ssllabs.com")
	/*
		if err != nil {
			log.Fatal(err)
		}
	*/

	// Run all the lab
	//static host -> host := "github.com"
	host := readHostFromUser()

	// checking the host with the test case
	/*
		if err := test.RunAllStates(host); err != nil {
			log.Fatal(err)
	*/

	fmt.Println("Starting SSL Labs assessment for:", host)

	// Step 1: start new assessment (ONLY ONCE)
	resp, err := analyze(host, true)
	if err != nil {
		panic(err)
	}

	// Step 2: poll until READY or ERROR
	for {
		fmt.Printf("Status: %s\n", resp.Status)

		if resp.Status == "READY" || resp.Status == "ERROR" {
			break
		}

		time.Sleep(10 * time.Second)

		resp, err = analyze(host, false)
		if err != nil {
			panic(err)
		}
	}

	// Step 3: read results
	fmt.Println("\nFinal results:\n")

	for _, ep := range resp.Endpoints {
		fmt.Printf("IP: %s\n", ep.IPAddress)

		if ep.StatusMessage != "Ready" {
			fmt.Printf("  Status: %s\n\n", ep.StatusMessage)
			continue
		}

		fmt.Printf("  Grade: %s\n", ep.Grade)

		if ep.HasWarnings {
			fmt.Println("  Warnings present")
		}

		if ep.Details != nil {
			fmt.Print("  Protocols: ")
			for _, p := range ep.Details.Protocols {
				fmt.Printf("%s %s  ", p.Name, p.Version)
			}
			fmt.Println()

			fmt.Printf("  Forward Secrecy: %s\n", yesNo(ep.Details.ForwardSecrecy > 0))
			fmt.Printf("  Heartbleed: %s\n", yesNo(!ep.Details.Heartbleed))
			fmt.Printf("  POODLE: %s\n", yesNo(!ep.Details.Poodle))
			fmt.Printf("  Logjam: %s\n", yesNo(!ep.Details.Logjam))

			if len(ep.Details.Cert.CommonNames) > 0 {
				fmt.Printf("  Cert CN: %s\n", ep.Details.Cert.CommonNames[0])
			}

			exp := time.Unix(ep.Details.Cert.NotAfter/1000, 0)
			fmt.Printf("  Cert expires: %s\n", exp.Format("2006-01-02"))
		}

		fmt.Println()
	}
}
