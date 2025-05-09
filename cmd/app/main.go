package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
)

func main() {
	// Define the -p flag for concurrency
	concurrency := flag.Int("p", 1, "Number of concurrent curl requests")
	bodyFile := flag.String("body-file", "", "Path to a JSON file to use as request body")
	flag.Parse()

	// Original args excluding custom flags
	curlArgs := filterCustomArgs(os.Args[1:], []string{"-p", "--body-file"})

	// If body file is provided, read it and inject as --data
	if *bodyFile != "" {
		jsonBody, err := ioutil.ReadFile(*bodyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read body file: %v\n", err)
			os.Exit(1)
		}

		// Check if user already provided -d or --data
		if hasDataFlag(curlArgs) {
			fmt.Fprintln(os.Stderr, "Warning: Both --body-file and -d/--data provided. --body-file will be ignored.")
		} else {
			curlArgs = append(curlArgs, "--data", string(jsonBody))
		}
	}

	if len(curlArgs) == 0 {
		fmt.Println("Usage: pcurl -p <concurrency> [--body-file file.json] [curl arguments]")
		os.Exit(1)
	}

	// Run concurrently
	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cmd := exec.Command("curl", curlArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Request %d failed: %v\n", i+1, err)
			}
		}(i)
	}
	wg.Wait()
}

func filterCustomArgs(args []string, skipFlags []string) []string {
	var result []string
	skipNext := false
	for i := 0; i < len(args); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		for _, flag := range skipFlags {
			if args[i] == flag {
				skipNext = true
				goto skip
			}
		}
		result = append(result, args[i])
	skip:
	}
	return result
}

func hasDataFlag(args []string) bool {
	for i := 0; i < len(args); i++ {
		if args[i] == "-d" || args[i] == "--data" {
			return true
		}
	}
	return false
}
