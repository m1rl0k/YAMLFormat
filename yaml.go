package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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
		// Find the line number and content of the error
		line, content := findErrorLine(string(data), err)
		log.Printf("Error parsing YAML in file %s (line %d: %q): %v\n", path, line, content, err)
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

func findErrorLine(data string, err error) (int, string) {
	var line, column int

	if syntaxErr, ok := err.(*yaml.SyntaxError); ok {
		line, column = syntaxErr.Line, syntaxErr.Column
	}

	lines := strings.Split(data, "\n")
	if line > 0 && line <= len(lines) {
		return line, lines[line-1]
	}

	return -1, ""
}

func updateYAMLNodeStyle(node *yaml.Node) {
	if node.Kind == yaml.MappingNode || node.Kind == yaml.SequenceNode {
		node.Style = yaml.FlowStyle
	}

	for i := 0; i < len(node.Content); i++ {
		updateYAMLNodeStyle(node.Content[i])
	}
}

func showDiff(path, original, formatted string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(original, formatted, false)

	if len(diffs) > 1 {
		fmt.Printf("Differences in file %s:\n", path)
		fmt.Println(dmp.DiffPrettyText(diffs))
	}
}

func suggestFixForLine(line string) string {
	fixed := strings.TrimSpace(line)

	if strings.Contains(fixed, ":") {
		parts := strings.SplitN(fixed, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// If the value contains a colon, assume it's an error and remove it
		if strings.Contains(value, ":") {
			value = strings.Replace(value, ":", "", 1)
		}

		fixed = fmt.Sprintf("%s: %s", key, value)
	} else {
		// If there's no colon, assume it's missing and add one
		fixed = fmt.Sprintf("%s:", fixed)
	}

	return fixed
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
