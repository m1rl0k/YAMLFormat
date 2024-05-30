package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

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
			changes, originalData, formattedData, err := formatYAMLFile(path)
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
			if len(originalData) > 0 && len(formattedData) > 0 {
				diff := generateDiff(originalData, formattedData)
				if len(diff) > 0 {
					fmt.Println("Changes suggested for", path+":")
					fmt.Println(diff)
				}
			}
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

func formatYAMLFile(path string) (int, []byte, []byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", path, err)
		return 0, nil, nil, err
	}
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		correctedData, changes, err := correctYAMLData(data)
		if err != nil {
			fmt.Println("Error correcting YAML data:", err)
			return 0, nil, nil, err
		}
		if changes {
			return 1, nil, correctedData, nil
		}
		return 0, data, nil, nil
	}
	formattedData, err := yaml.Marshal(removeEmptyNodes(yamlData))
	if err != nil {
		fmt.Println("Error formatting YAML data:", err)
		return 0, nil, nil, err
	}
	expectedData := []byte(strings.TrimSpace(string(formattedData)))
	actualData := []byte(strings.TrimSpace(string(data)))
	if !bytes.Equal(expectedData, actualData) {
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			ToFile:   path,
			Context:  3,
		}
		text, err := difflib.GetUnifiedDiffString(diff)
		if err != nil {
			fmt.Println("Error generating diff:", err)
			return 0, nil, nil, err
		}
		return 1, actualData, []byte(text), nil
	}
	fmt.Printf("\033[32mNo changes needed for %s\n\033[0m", path)
	return 0, data, nil, nil
}

func correctYAMLData(data []byte) ([]byte, bool, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, false, err
	}
	corrected := traverseYAMLTree(yamlData)
	if corrected {
		correctedData, err := yaml.Marshal(yamlData)
		if err != nil {
			return nil, false, err
		}
		return correctedData, true, nil
	}
	return data, false, nil
}

func traverseYAMLTree(node interface{}) bool {
	switch node := node.(type) {
	case map[string]interface{}:
		for key, value := range node {
			if mapValue, ok := value.(map[string]interface{}); ok {
				if traverseYAMLTree(mapValue) {
					node[key] = mapValue
				}
			} else if listValue, ok := value.([]interface{}); ok {
				if traverseYAMLList(listValue) {
					node[key] = listValue
				}
			} else {
				switch value.(type) {
				case map[interface{}]interface{}:
					mapValue := make(map[string]interface{})
					for k, v := range value.(map[interface{}]interface{}) {
						mapValue[fmt.Sprintf("%v", k)] = v
					}
					node[key] = mapValue
					return true
				case string:
					if newValue := correctString(value.(string)); newValue != value {
						node[key] = newValue
						return true
					}
				case nil:
					delete(node, key)
					return true
				}
			}
		}
	case []interface{}:
		if traverseYAMLList(node) {
			return true
		}
	}
	return false
}

func traverseYAMLList(list []interface{}) bool {
	changed := false
	for i, item := range list {
		if mapValue, ok := item.(map[string]interface{}); ok {
			if traverseYAMLTree(mapValue) {
				list[i] = mapValue
				changed = true
			}
		} else if nestedList, ok := item.([]interface{}); ok {
			if traverseYAMLList(nestedList) {
				list[i] = nestedList
				changed = true
			}
		} else {
			switch item.(type) {
			case map[interface{}]interface{}:
				mapValue := make(map[string]interface{})
				for k, v := range item.(map[interface{}]interface{}) {
					mapValue[fmt.Sprintf("%v", k)] = v
				}
				list[i] = mapValue
				changed = true
			}
		}
	}
	return changed
}

func removeEmptyNodes(node interface{}) interface{} {
	switch node := node.(type) {
	case map[string]interface{}:
		for key, value := range node {
			node[key] = removeEmptyNodes(value)
		}
		for key, value := range node {
			if isEmpty(value) {
				delete(node, key)
			}
		}
	case []interface{}:
		for i := range node {
			node[i] = removeEmptyNodes(node[i])
		}
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

func generateDiff(originalData, correctedData []byte) string {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(originalData)),
		B:        difflib.SplitLines(string(correctedData)),
		FromFile: "Original",
		ToFile:   "Formatted",
		Context:  3,
	}
	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		fmt.Println("Error generating diff:", err)
		return ""
	}
	lines := strings.Split(text, "\n")
	var buf bytes.Buffer
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+"):
			buf.WriteString("\033[32m")
		case strings.HasPrefix(line, "-"):
			buf.WriteString("\033[31m")
		case strings.HasPrefix(line, "@"):
			buf.WriteString("\033[36m")
		default:
			buf.WriteString("\033[0m")
		}
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	return buf.String()
}
