package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v3"
)

func main() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		os.Exit(1)
	}

	totalChanges := 0

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") || strings.HasSuffix(f.Name(), ".yml") {
			changes := processYAMLFile(f.Name())
			totalChanges += changes
		}
	}

	fmt.Printf("\n\033[1mTotal changes: %d\033[0m\n", totalChanges)
}

func processYAMLFile(filename string) int {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", filename, err)
		return 0
	}

	var parsedData interface{}
	err = yaml.Unmarshal(data, &parsedData)
	if err != nil {
		fmt.Println("Error parsing file:", filename, err)
		return 0
	}

	formattedData, err := yaml.Marshal(parsedData)
	if err != nil {
		fmt.Println("Error formatting file:", filename, err)
		return 0
	}

	if string(data) != string(formattedData) {
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			FromFile: "Original",
			ToFile:   "Formatted",
			Context:  3,
		}
		diffStr, _ := difflib.GetUnifiedDiffString(diff)

		fmt.Printf("\n\033[33mProposed changes for %s:\033[0m\n", filename)
		fmt.Println(strings.Repeat("=", 50))
		fmt.Println(formatDiff(diffStr))
		fmt.Println(strings.Repeat("=", 50))

		return countChanges(diffStr)
	} else {
		fmt.Printf("\033[32mNo changes needed for %s\n\033[0m", filename)
		return 0
	}
}

func formatDiff(diff string) string {
	var formattedDiff strings.Builder

	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "+"):
			formattedDiff.WriteString("\033[32m" + line + "\033[0m")
		case strings.HasPrefix(line, "-"):
			formattedDiff.WriteString("\033[31m" + line + "\033[0m")
		default:
			formattedDiff.WriteString(line)
		}

		formattedDiff.WriteString("\n")
	}

	return formattedDiff.String()
}

func countChanges(diff string) int {
	count := 0
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			if !strings.HasPrefix(line, "+++") && !strings.HasPrefix(line, "---") {
				count++
			}
		}
	}
	return count
}
