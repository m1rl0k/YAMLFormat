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
	
	// Attempt to parse the YAML data
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		// Handle the error by applying corrections to the data
		correctedData, corrected, err := correctYAMLData(data)
		if err != nil {
			fmt.Println("Error correcting YAML data:", err)
			return 0, nil, err
		}
		
		// Return the corrected data and the number of changes made
		if corrected {
			if err := ioutil.WriteFile(path, correctedData, 0644); err != nil {
				fmt.Println("Error writing corrected file", path, ":", err)
				return 0, nil, err
			}
			return 1, correctedData, nil
		}
		
		// Return the original data if no corrections were made
		return 0, data, nil
	}
	
	// If the data was successfully parsed, reformat it and compare to the original
	formattedData, err := yaml.Marshal(yamlData)
	if err != nil {
		fmt.Println("Error formatting YAML data:", err)
		return 0, nil, err
	}
	if !bytes.Equal(data, formattedData) {
		// Generate a unified diff of the changes
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

// Attempt to correct YAML data errors
func correctYAMLData(data []byte) ([]byte, bool, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, false, err
	}
	
	// Recursively traverse the YAML tree and correct any formatting errors
	corrected := traverseYAMLTreeCorrect(yamlData)
	
	// If corrections were made, reformat the YAML data and return it
	if corrected {
		correctedData, err := yaml.Marshal(yamlData)
		if err != nil {
			return nil, false, err
		}
		return correctedData, true, nil
	}
	
	// If no corrections were made, return the original data
	return data, false, nil
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
func traverseYAMLTreeCorrect(node interface{}) {
	switch node := node.(type) {
	case map[interface{}]interface{}:
		for key, value := range node {
			if mapValue, ok := value.(map[interface{}]interface{}); ok {
				// recursively traverse nested map
				traverseYAMLTreeCorrect(mapValue)
			} else if listValue, ok := value.([]interface{}); ok {
				// recursively traverse nested list
				traverseYAMLListCorrect(listValue)
			}
			if keyStr, ok := key.(string); ok {
				// correct key if necessary
				if strings.Contains(keyStr, " ") {
					delete(node, key)
					node[strings.ReplaceAll(keyStr, " ", "_")] = value
				}
			}
		}
	case []interface{}:
		traverseYAMLListCorrect(node)
	}
}

func traverseYAMLListCorrect(node []interface{}) {
	for _, value := range node {
		if mapValue, ok := value.(map[interface{}]interface{}); ok {
			// recursively traverse nested map
			traverseYAMLTreeCorrect(mapValue)
		} else if listValue, ok := value.([]interface{}); ok {
			// recursively traverse nested list
			traverseYAMLListCorrect(listValue)
		}
	}
}

func removeEmptyNodes(node interface{}) interface{} {
    switch node := node.(type) {
    case map[string]interface{}:
        for key, value := range node {
            node[key] = removeEmptyNodes(value)
        }
        // remove any keys with empty values
        for key, value := range node {
            if isEmpty(value) {
                delete(node, key)
            }
        }
    case []interface{}:
        for i := range node {
            node[i] = removeEmptyNodes(node[i])
        }
        // remove any empty nodes from list
        for i := 0; i < len(node); i++ {
            if isEmpty(node[i]) {
                node = append(node[:i], node[i+1:]...)
                i--
            }
        }
    }
    return node
}

func isEmpty(node interface{}) bool {
    switch node := node.(type) {
    case map[string]interface{}:
        return len(node) == 0
    case []interface{}:
        return len(node) == 0
    case string:
        return node == ""
    default:
        return false
    }
}

func correctString(str string) string {
    str = strings.TrimSpace(str)
    str = strings.ReplaceAll(str, "\\n", "\n")
    return str
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
