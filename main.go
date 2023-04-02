package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
        "bytes"


	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v3"
        


)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	totalChanges := 0
	var changedFiles []string

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".yaml" || ext == ".yml" {
			changes, formattedData, err := formatYAMLFile(path)
			if err != nil {
				fmt.Println("Error formatting file", path, ":", err)
				return nil
			}
			if changes > 0 {
				changedFiles = append(changedFiles, path)
				if err := ioutil.WriteFile(path, formattedData, info.Mode()); err != nil {
					fmt.Println("Error writing formatted file", path, ":", err)
					return nil
				}
			}
			totalChanges += changes
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if totalChanges > 0 {
		fmt.Printf("\n\033[1mTotal changes: %d\033[0m\n", totalChanges)
		fmt.Println("The following files have changes:")
		for _, file := range changedFiles {
			fmt.Printf("- %s\n", file)
		}
	} else {
		fmt.Println("\n\033[32mNo changes needed\033[0m")
	}
}


func formatYAMLFile(path string) (int, []byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", path, err)
		return 0, nil, err
	}
	formattedData, err := formatYAML(data)
	if err != nil {
		fmt.Println("Error formatting YAML data:", err)
		return 0, nil, err
	}
	if !bytes.Equal(data, formattedData) {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			FromFile: "Original",
			ToFile:   "Formatted",
			Context:  3,
		})

		fmt.Printf("\n\033[33mProposed changes for %s:\033[0m\n", path)
		fmt.Println(strings.Repeat("=", 50))
		fmt.Println(formatDiff(diff))
		fmt.Println(strings.Repeat("=", 50))

		return countChanges(diff), formattedData, nil
	}

	fmt.Printf("\033[32mNo changes needed for %s\n\033[0m", path)
	return 0, nil, nil
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

func formatYAML(data []byte) ([]byte, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, err
	}

	traverseYAMLTree(yamlData)

	formattedData, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, err
	}

	return formattedData, nil
}

func traverseYAMLTree(node interface{}) {
	switch node := node.(type) {
	case map[string]interface{}:
		for key, value := range node {
			if mapValue, ok := value.(map[string]interface{}); ok {
				// recursively traverse nested map
				traverseYAMLTree(mapValue)
			} else if listValue, ok := value.([]interface{}); ok {
				// recursively traverse nested list
				traverseYAMLList(listValue)
			}
		}
	case []interface{}:
		traverseYAMLList(node)
	}
}

func traverseYAMLList(list []interface{}) {
	for _, value := range list {
		if mapValue, ok := value.(map[string]interface{}); ok {
			// recursively traverse nested map
			traverseYAMLTree(mapValue)
		} else if listValue, ok := value.([]interface{}); ok {
			// recursively traverse nested list
			traverseYAMLList(listValue)
		}
	}
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
