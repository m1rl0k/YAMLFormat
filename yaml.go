package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

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

func showDiff(path, original, formatted string) {
	var builder strings.Builder
	originalLines := strings.Split(original, "\n")
	formattedLines := strings.Split(formatted, "\n")

	for i := 0; i < len(originalLines) || i < len(formattedLines); i++ {
		var originalLine, formattedLine string
		if i < len(originalLines) {
			originalLine = originalLines[i]
		}
		if i < len(formattedLines) {
			formattedLine = formattedLines[i]
		}

		switch {
		case originalLine == formattedLine:
			builder.WriteString(originalLine)
			builder.WriteString("\n")
		case len(originalLine) == 0:
			builder.WriteString("\x1b[32m" + formattedLine + "\x1b[0m")
			builder.WriteString("\n")
		case len(formattedLine) == 0:
			builder.WriteString("\x1b[31m" + originalLine + "\x1b[0m")
			builder.WriteString("\n")
		default:
			start := 0
			for start < len(originalLine) && start < len(formattedLine) && originalLine[start] == formattedLine[start] {
				start++
			}
			end := 0
			for end < len(originalLine)-start && end < len(formattedLine)-start && originalLine[len(originalLine)-end-1] == formattedLine[len(formattedLine)-end-1] {
				end++
			}
			if start > 0 || end > 0 {
				builder.WriteString(originalLine[:start])
			}
			builder.WriteString("\x1b[31m" + originalLine[start:len(originalLine)-end] + "\x1b[0m")
			builder.WriteString("\x1b[32m" + formattedLine[start:len(formattedLine)-end] + "\x1b[0m")
			if start > 0 || end > 0 {
				builder.WriteString(formattedLine[len(formattedLine)-end:])
			}
			builder.WriteString("\n")
		}
	}

	diff := builder.String()
	if diff != "" {
		fmt.Printf("Differences in file %s:\n%s", path, diff)
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

	// Check for missing quotes around strings
	re := regexp.MustCompile(`[^#]\s*\S*:\s*\S+`)
	matches := re.FindAllStringIndex(fixedLine, -1)
	for _, match := range matches {
		value := fixedLine[match[1]:]
		if !strings.HasPrefix(value, "'") && !strings.HasPrefix(value, "\"") {
			valueStart := strings.IndexFunc(value, func(r rune) bool { return !unicode.IsSpace(r) })
			if valueStart != -1 {
				fixedLine = fixedLine[:match[1]+valueStart] + "\"" + value[valueStart:] + "\""
			}
		}
	}

	return fixedLine
}
