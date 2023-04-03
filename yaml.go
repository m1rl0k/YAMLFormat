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
	"unicode"

	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

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
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", path, err)
		return
	}

	var parsedContent yaml.Node
	err = yaml.Unmarshal(data, &parsedContent)
	if err != nil {
		log.Printf("Error parsing YAML in file %s: %v\n", path, err)
		return
	}

	updateYAMLNodeStyle(&parsedContent)

	formatted, err := yaml.Marshal(&parsedContent)
	if err != nil {
		log.Printf("Error formatting YAML in file %s: %v\n", path, err)
		return
	}

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
		pos := 0
		for _, delta := range deltas {
			deltaType := string(delta[0])

			switch deltaType {
			case "+":
				length, _ := strconv.Atoi(string(delta[1:]))
				fmt.Println("Added:")
				fmt.Println(delta[2:])
				pos += length
			case "-":
				length, _ := strconv.Atoi(string(delta[1:]))
				fmt.Println("Removed:")
				start := pos - 12
				if start < 0 {
					start = 0
				}
				end := pos + length + 12
				if end >= len(original) {
					end = len(original) - 1
				}

				fmt.Println("Original:")
				for i := start; i <= end; i++ {
					fmt.Printf("%4d: %s\n", i+1, string(original[i]))
				}

				fmt.Println("Formatted:")
				for i := start; i <= end; i++ {
					fmt.Printf("%4d: %s\n", i+1, string(formatted[i]))
				}

				pos += length
			case "=":
				length, _ := strconv.Atoi(string(delta[1:]))
				pos += length
			}
		}
	}
}


func suggestFixForLine(line string) string {
	fixedLine := line

	// Check for missing colons
	if !strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
		index := strings.IndexFunc(line, unicode.IsLetter)
		if index != -1 {
			fixedLine = line[:index+1] + ": " + line[index+1:]
		}
	}

	// Check for missing spaces after colons
	colonIndex := strings.Index(line, ":")
	if colonIndex != -1 && len(line) > colonIndex+1 && line[colonIndex+1] != ' ' {
		fixedLine = line[:colonIndex+1] + " " + line[colonIndex+1:]
	}

	// Fix incorrect indentation
	indentation := 0
	for _, char := range line {
		if char == ' ' {
			indentation++
		} else {
			break
		}
	}

	correctIndentation := (indentation / 2) * 2
	if correctIndentation != indentation {
		fixedLine = strings.Repeat(" ", correctIndentation) + strings.TrimSpace(line)
	}

	return fixedLine
}
