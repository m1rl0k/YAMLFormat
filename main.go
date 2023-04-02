package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v3"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".yaml" || ext == ".yml" {
			return formatYAMLFile(path)
		}
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func formatYAMLFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	formattedData, err := formatYAML(data)
	if err != nil {
		return err
	}
	if !strings.EqualFold(string(data), string(formattedData)) {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			FromFile: "Original",
			ToFile:   "Formatted",
			Context:  3,
		})
		fmt.Printf("\033[1;33mDifferences in YAML file: %s\033[0m\n", path)
		fmt.Println(formatDiff(diff))
		printSummary(diff)
	}
	return nil
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

func printSummary(diff string) {
	added := 0
	removed := 0
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") {
			added++
		} else if strings.HasPrefix(line, "-") {
			removed++
		}
	}

	fmt.Printf("\033[1;36mSummary:\033[0m\n")
	fmt.Printf("\033[32mAdded lines: %d\033[0m\n", added)
	fmt.Printf("\033[31mRemoved lines: %d\033[0m\n\n", removed)
}

func formatYAML(data []byte) ([]byte, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, err
	}

	formattedData, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, err
	}

	return formattedData, nil
}
