package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

func main() {
	// Define the -p flag for concurrency
	concurrency := flag.Int("p", 1, "Number of concurrent curl requests")
	flag.Parse()

	// Get all the other args meant for curl (excluding -p and its value)
	curlArgs := filterCurlArgs(os.Args[1:], "-p")

	if len(curlArgs) == 0 {
		fmt.Println("Usage: pcurl -p <concurrency> [curl arguments]")
		os.Exit(1)
	}

	// Run the curl command in parallel
	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Printf("CURL %d started at: %v\n", i+1, time.Now().Format(time.RFC3339))
			cmd := exec.Command("curl", curlArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Request %d failed: %v\n", i+1, err)
			}
			fmt.Printf("CURL %d finished at:%v\n", i+1, time.Now().Format(time.RFC3339))
		}(i)
	}

	wg.Wait()
}

func filterCurlArgs(args []string, skipFlag string) []string {
	var result []string
	skipNext := false
	for i := 0; i < len(args); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		if args[i] == skipFlag {
			skipNext = true
			continue
		}
		result = append(result, args[i])
	}
	return result
}
