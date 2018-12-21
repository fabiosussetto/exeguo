package utils

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

func formatCmdOutput(rawOutput string) string {
	return strings.TrimSuffix(rawOutput, "\n")
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

// Run parallelise command exec
func Run() []string {
	var wg sync.WaitGroup
	inputRange := makeRange(0, 1000)

	c := make(chan string, len(inputRange))

	for index := range inputRange {
		wg.Add(1)
		go getCurrentDate(&wg, c, index)
	}

	wg.Wait()
	close(c)

	var results []string

	for resMsg := range c {
		results = append(results, resMsg)
	}

	return results
}

func getCurrentDate(wg *sync.WaitGroup, c chan<- string, index int) {
	defer wg.Done()

	dateCmd := exec.Command("date")

	dateOut, err := dateCmd.Output()

	if err != nil {
		panic(err)
	}

	parsedOut := formatCmdOutput(string(dateOut))

	result := fmt.Sprintf("Yo %s!, [%d]", parsedOut, index)

	c <- result
}
