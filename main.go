package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
        "regexp"
	"strconv"

	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"
)

func main() {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Read and process all YAML files in the directory
	err = filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			processYAMLFile(path)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func processYAMLFile(path string) {
	// Read the YAML file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", path, err)
		return
	}

	// Parse the YAML content
	var parsedContent yaml.Node
	err = yaml.Unmarshal(data, &parsedContent)
	if err != nil {
		log.Printf("Error parsing YAML in file %s: %v\n", path, err)
		return
	}

	// Update YAML node style
	updateYAMLNodeStyle(&parsedContent)

	// Format the parsed YAML content
	formatted, err := yaml.Marshal(&parsedContent)
	if err != nil {
		log.Printf("Error formatting YAML in file %s: %v\n", path, err)
		return
	}

	// Generate and display the diff
	showDiff(path, string(data), string(formatted))
}

func updateYAMLNodeStyle(node *yaml.Node) {
	if node.Kind == yaml.MappingNode || node.Kind == yaml.SequenceNode {
		node.Style = yaml.FlowStyle
	}

	for i := 0; i < len(node.Content); i++ {
		updateYAMLNodeStyle(node.Content[i])
	}
}

func findErrorLineAndSuggestFix(data string, err error) (int, string, string) {
	line := -1

	// Use a regex to extract the line number from the error message
	re := regexp.MustCompile(`line (\d+):`)
	matches := re.FindStringSubmatch(err.Error())
	if len(matches) > 1 {
		var err error
		line, err = strconv.Atoi(matches[1])
		if err != nil {
			line = -1
		}
	}

	lines := strings.Split(data, "\n")
	if line > 0 && line <= len(lines) {
		return line, lines[line-1], suggestFixForLine(lines[line-1])
	}

	return -1, "", ""
}

func showDiff(path, original, formatted string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(original, formatted, false)

	if len(diffs) > 1 {
		fmt.Printf("Differences in file %s:\n", path)

		deltas := dmp.DiffToDelta(diffs)
		originalLines := strings.Split(original, "\n")
		formattedLines := strings.Split(formatted, "\n")

		pos := 0
		for _, delta := range deltas {
			if delta.Type == diffmatchpatch.DiffEqual {
				pos += delta.Length1
				continue
			}

			start := pos - 12
			if start < 0 {
				start = 0
			}
			end := pos + 12
			if end >= len(originalLines) {
				end = len(originalLines) - 1
			}

			fmt.Println("Original:")
			for i := start; i <= end; i++ {
				fmt.Printf("%4d: %s\n", i+1, originalLines[i])
			}

			fmt.Println("Formatted:")
			for i := start; i <= end; i++ {
				fmt.Printf("%4d: %s\n", i+1, formattedLines[i])
			}

			pos += delta.Length1
		}
	}
}

func findErrorLineAndSuggestFix(data string, err error) (int, string, string) {
	var line, column int

	if syntaxErr, ok := err.(*yaml.SyntaxError); ok {
		line, column = syntaxErr.Line, syntaxErr.Column
	}

	lines := strings.Split(data, "\n")
	if line > 0 && line <= len(lines) {
		return line, lines[line-1], suggestFixForLine(lines[line-1])
	}

	return -1, "", ""
}
