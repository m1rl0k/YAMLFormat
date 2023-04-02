package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"
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
		// Find the line number, content, and suggested fix of the error
		line, content, suggestedFix := findErrorLineAndSuggestFix(string(data), err)
		log.Printf("Error parsing YAML in file %s (line %d: %q): %v\n", path, line, content, err)
		log.Printf("Suggested fix for line %d: %q\n", line, suggestedFix)
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

func findErrorLineAndSuggestFix(data string, err error) (int, string, string) {
	line := -1

	// Use a regex to extract the line number from the error message
	re := regexp.MustCompile(`line (\d+):`)
	matches := re.FindStringSubmatch(err.Error())
	if len(matches) > 1 {
		var convErr error
		line, convErr = strconv.Atoi(matches[1])
		if convErr != nil {
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
		coloredOutput := dmp.DiffPrettyHtml(diffs)
		colorable.NewColorableStdout().Write([]byte(coloredOutput))
		fmt.Println()
	}
}

func showDiff(path, original, formatted string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(original, formatted, false)

	if len(diffs) > 1 {
		fmt.Printf("Differences in file %s:\n", path)
		coloredOutput := diffToColoredYAMLText(diffs)
		colorable.NewColorableStdout().Write([]byte(coloredOutput))
		fmt.Println()
	}
}

func diffToColoredYAMLText(diffs []diffmatchpatch.Diff) string {
	var output strings.Builder
	for _, diff := range diffs {
		text := diff.Text

		if diff.Type == diffmatchpatch.DiffInsert {
			// Green for inserted lines
			output.WriteString("\x1b[32m")
			output.WriteString(text)
			output.WriteString("\x1b[0m")
		} else if diff.Type == diffmatchpatch.DiffDelete {
			// Red for deleted lines
			output.WriteString("\x1b[31m")
			output.WriteString(text)
			output.WriteString("\x1b[0m")
		} else {
			output.WriteString(text)
		}
	}
	return output.String()
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
