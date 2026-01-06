package test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RunAllStates starts all the process and print
// All the json entirely
func RunAllStates(host string) error {
	startURL := fmt.Sprintf(
		"https://api.ssllabs.com/api/v2/analyze?host=%s&startNew=on&all=done",
		host,
	)

	checkURL := fmt.Sprintf(
		"https://api.ssllabs.com/api/v2/analyze?host=%s&all=done",
		host,
	)

	// Start assessment
	_, _, err := printJSON(startURL)
	if err != nil {
		return err
	}

	//Poll until READY or ERROR
	for {
		time.Sleep(10 * time.Second)

		status, done, err := printJSON(checkURL)
		if err != nil {
			return err
		}

		if done {
			fmt.Println("Assessment finished with status:", status)
			fmt.Println("Test ended")
			break
		}
	}

	return nil
}

func printJSON(url string) (status string, done bool, err error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}

	fmt.Println("====== JSON RESPONSE ======")
	fmt.Println(string(body))
	fmt.Println("===========================")

	if strings.Contains(string(body), `"status":"READY"`) {
		return "READY", true, nil
	}
	if strings.Contains(string(body), `"status":"ERROR"`) {
		return "ERROR", true, nil
	}

	return "", false, nil
}
