package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
        "regexp"
        "strconv"
        
        "github.com/ghodss/yaml"
	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v3"

)

func main() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		os.Exit(1)
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") || strings.HasSuffix(f.Name(), ".yml") {
			lintYAMLFile(f.Name())
		}
	}
}

func lintYAMLFile(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", filename, err)
		return
	}

	var parsedData interface{}
	err = yaml.Unmarshal(data, &parsedData)
	if err != nil {
		fmt.Printf("\033[31mSyntax issues found in %s:\033[0m\n", filename)
		fmt.Println(err)
		provideSuggestions(string(data), err)
	} else {
		fmt.Printf("\033[32mNo syntax issues found in %s\n\033[0m", filename)
	}
}

func provideSuggestions(data string, yamlErr error) {
	lines := strings.Split(data, "\n")

	re := regexp.MustCompile(`line (\d+):`)
	matches := re.FindStringSubmatch(yamlErr.Error())
	if len(matches) > 1 {
		lineNum := matches[1]
		fmt.Printf("\033[33mSuggestion for line %s:\033[0m\n", lineNum)

		// Common issues
		fmt.Println("1. Check indentation, use two spaces for each level.")
		fmt.Println("2. Ensure there is a space after colons.")
		fmt.Println("3. Verify that there are no tabs or extra spaces before or after keys and values.")
		fmt.Println("4. Check for missing or misplaced quotes for string values.")

		// Display the affected line
		line, err := strconv.Atoi(lineNum)
		if err == nil && line-1 < len(lines) {
			fmt.Printf("\033[34mAffected line:\033[0m\n%s\n", lines[line-1])
		}
	} else {
		fmt.Println("\033[33mGeneral suggestions:\033[0m")
		fmt.Println("1. Check the entire YAML file for correct syntax and structure.")
		fmt.Println("2. Use a YAML linter or validator to identify and fix specific issues.")
	}
}

