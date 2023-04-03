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
        "github.com/pmezard/go-difflib/difflib"
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
	chars1, chars2, lines := dmp.DiffLinesToChars(original, formatted)
	diffs := dmp.DiffMain(chars1, chars2, false)
	diffs = dmp.DiffCharsToLines(diffs, lines)

	if len(diffs) > 1 {
		fmt.Printf("Differences in file %s:\n", path)

		var addedLines []string
		var removedLines []string

		for _, diff := range diffs {
			switch diff.Type {
			case diffmatchpatch.DiffInsert:
				addedLines = append(addedLines, diff.Text)
			case diffmatchpatch.DiffDelete:
				removedLines = append(removedLines, diff.Text)
			}
		}

		if len(addedLines) > 0 {
			fmt.Println("Added:")
			for _, line := range addedLines {
				fmt.Print(line)
			}
		}

		if len(removedLines) > 0 {
			fmt.Println("Removed:")
			for _, line := range removedLines {
				fmt.Print(line)
			}
		}

		if len(addedLines) > 0 && len(removedLines) > 0 {
			fmt.Println("Changes:")
			diffs = dmp.DiffMain(strings.Join(removedLines, ""), strings.Join(addedLines, ""), false)
			for _, diff := range diffs {
				switch diff.Type {
				case diffmatchpatch.DiffEqual:
					fmt.Print(diff.Text)
				case diffmatchpatch.DiffInsert:
					fmt.Printf("\x1b[32m%s\x1b[0m", diff.Text)
				case diffmatchpatch.DiffDelete:
					fmt.Printf("\x1b[31m%s\x1b[0m", diff.Text)
				}
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
