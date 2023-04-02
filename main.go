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

func suggestFixForLine(line string) string {
	fixed := strings.TrimSpace(line)

	if strings.HasPrefix(fixed, "-") {
		// Fix misplaced hyphens for list items
		hyphenIndex := strings.Index(line, "-")
		if hyphenIndex > 0 {
			fixed = strings.Repeat(" ", hyphenIndex) + "- " + strings.TrimSpace(fixed[1:])
		} else {
			fixed = "- " + strings.TrimSpace(fixed[1:])
		}
	} else if strings.Contains(fixed, ":") {
		// Fix extra colons or incorrect indentation for key-value pairs
		parts := strings.SplitN(fixed, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// If the value contains a colon, assume it's an error and remove it
		if strings.Contains(value, ":") {
			value = strings.Replace(value, ":", "", 1)
		}

		fixed = fmt.Sprintf("%s: %s", key, value)
	} else {
		// Assume a missing colon for other cases
		fixed = fmt.Sprintf("%s:", fixed)
	}

	return fixed
}

func showDiff(path, original, formatted string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(original, formatted, false)

	if len(diffs) > 1 {
		fmt.Printf("Differences in file %s:\n", path)
		fmt.Println(dmp.DiffPrettyText(diffs))
	}
}
