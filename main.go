package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
        "bytes"


	
	"gopkg.in/yaml.v3"
        "github.com/pmezard/go-difflib/difflib"
       
        


)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var changedFiles []string
	var changes [][]byte

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".yaml" || ext == ".yml" {
			before, after, err := formatYAMLFile(path)
			if err != nil {
				fmt.Println("Error formatting file", path, ":", err)
				return nil
			}
			if after != nil {
				changedFiles = append(changedFiles, path)
				changes = append(changes, after)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(changes) > 0 {
		fmt.Printf("\n\033[1mChanges suggested:\033[0m\n")
		for i, file := range changedFiles {
			fmt.Printf("\033[1m%s:\033[0m\n%s\n", file, changes[i])
		}
	} else {
		fmt.Println("\n\033[32mNo changes needed\033[0m")
	}
}


func formatYAMLFile(path string) (int, []byte, []byte, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", path, err)
		return nil, nil, err
	}

	// Attempt to parse the YAML data
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		// Handle the error by applying corrections to the data
		correctedData, corrected, err := correctYAMLData(data)
		if err != nil {
			fmt.Println("Error correcting YAML data:", err)
			return nil, nil, err
		}

		// Return the corrected data and the number of changes made
		if corrected {
			return nil, correctedData, nil
		}

		// Return the original data if no corrections were made
		return data, nil, nil
	}

	// If the data was successfully parsed, reformat it and compare to the original
	formattedData, err := yaml.Marshal(removeEmptyNodes(yamlData))
	if err != nil {
		fmt.Println("Error formatting YAML data:", err)
		return nil, nil, err
	}

	// Check if the indentation is correct
	expectedData := []byte(strings.TrimSpace(string(formattedData)))
	actualData := []byte(strings.TrimSpace(string(data)))
	if !bytes.Equal(expectedData, actualData) {
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			FromFile: "Original",
			ToFile:   "Formatted",
			Context:  3,
		}
		text, err := difflib.GetUnifiedDiffString(diff)
		if err != nil {
			fmt.Println("Error generating diff:", err)
			return nil, nil, err
		}
		fmt.Printf("\033[33mChanges suggested for %s:\n\033[0m", path)
		fmt.Println(text)

		return actualData, []byte(text), nil
	}

	fmt.Printf("\033[32mNo changes needed for %s\n\033[0m", path)
	return data, nil, nil
}


func fixIndentation(expectedData, actualData []byte) ([]byte, error) {
    // Use a YAML library to parse both the expected and actual data
    var expectedYaml, actualYaml interface{}
    if err := yaml.Unmarshal(expectedData, &expectedYaml); err != nil {
        return nil, err
    }
    if err := yaml.Unmarshal(actualData, &actualYaml); err != nil {
        return nil, err
    }
    
    // Generate a map of expected node paths to indentation levels
    pathIndentMap := make(map[string]int)
    generatePathIndentMap(expectedYaml, pathIndentMap, 0)
    
    // Use the pathIndentMap to generate the corrected data
    correctedData := generateCorrectedData(actualYaml, pathIndentMap, 0)
    return correctedData, nil
}

func generatePathIndentMap(node interface{}, pathIndentMap map[string]int, indentLevel int) {
    switch node := node.(type) {
    case map[interface{}]interface{}:
        for k, v := range node {
            path := fmt.Sprintf("%s.%s", k, getTypeName(v))
            pathIndentMap[path] = indentLevel
            generatePathIndentMap(v, pathIndentMap, indentLevel+1)
        }
    case []interface{}:
        for i, v := range node {
            path := fmt.Sprintf("[%d].%s", i, getTypeName(v))
            pathIndentMap[path] = indentLevel
            generatePathIndentMap(v, pathIndentMap, indentLevel+1)
        }
    }
}

func traverseYAMLTree(node interface{}) bool {
    switch node := node.(type) {
    case map[string]interface{}:
        for key, value := range node {
            if mapValue, ok := value.(map[string]interface{}); ok {
                // recursively traverse nested map
                if traverseYAMLTree(mapValue) {
                    node[key] = mapValue
                }
            } else if listValue, ok := value.([]interface{}); ok {
                // recursively traverse nested list
                if traverseYAMLList(listValue) {
                    node[key] = listValue
                }
            } else {
                // handle case where value is not a map or list
                switch value.(type) {
                case map[interface{}]interface{}:
                    // convert map[interface{}]interface{} to map[string]interface{}
                    mapValue := make(map[string]interface{})
                    for k, v := range value.(map[interface{}]interface{}) {
                        mapValue[fmt.Sprintf("%v", k)] = v
                    }
                    node[key] = mapValue
                    return true
                case string:
                    // attempt to correct string value
                    if newValue := correctString(value.(string)); newValue != value {
                        node[key] = newValue
                        return true
                    }
                case nil:
                    // delete key with nil value
                    delete(node, key)
                    return true
                }
            }
        }
    case []interface{}:
        // recursively traverse list
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
            // recursively traverse nested map
            if traverseYAMLTree(mapValue) {
                list[i] = mapValue
                changed = true
            }
        } else if nestedList, ok := item.([]interface{}); ok {
            // recursively traverse nested list
            if traverseYAMLList(nestedList) {
                list[i] = nestedList
                changed = true
            }
        } else {
            // handle case where value is not a map or list
            switch item.(type) {
            case map[interface{}]interface{}:
                // convert map[interface{}]interface{} to map[string]interface{}
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

func correctYAMLData(data []byte) ([]byte, bool, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, false, err
	}
	
	// Recursively traverse the YAML tree and correct any formatting errors
	corrected := traverseYAMLTree(yamlData)
	
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

func getTypeName(v interface{}) string {
	switch v.(type) {
	case map[interface{}]interface{}, []interface{}:
		return "map"
	case string:
		return "string"
	case bool:
		return "bool"
	case int, int8, int16, int32, int64:
		return "int"
	case uint, uint8, uint16, uint32, uint64:
		return "uint"
	case float32, float64:
		return "float"
	default:
		return "unknown"
	}
}
func generateCorrectedData(actualData interface{}, pathIndentMap map[string]int, indentLevel int) []byte {
	var buf bytes.Buffer

	switch actualData := actualData.(type) {
	case map[interface{}]interface{}:
		buf.WriteString("{\n")
		for k, v := range actualData {
			key := fmt.Sprintf("%v", k)
			value := getTypeValue(v)

			indentStr := strings.Repeat("  ", indentLevel+1)
			path := fmt.Sprintf("%s.%s", key, getTypeName(v))

			if indent, ok := pathIndentMap[path]; ok {
				buf.WriteString(strings.Repeat("  ", indent+1))
			} else {
				buf.WriteString(indentStr)
			}

			buf.WriteString(fmt.Sprintf("%s: %s", key, value))

			if indent, ok := pathIndentMap[path]; ok && indentLevel+1 <= indent {
				buf.WriteString("\n")
			} else {
				buf.WriteString(",\n")
			}

			generateCorrectedData(v, pathIndentMap, indentLevel+1)
		}
		buf.WriteString(strings.Repeat("  ", indentLevel))
		buf.WriteString("}")
	case []interface{}:
		buf.WriteString("[\n")
		for i, v := range actualData {
			value := getTypeValue(v)

			indentStr := strings.Repeat("  ", indentLevel+1)
			path := fmt.Sprintf("[%d].%s", i, getTypeName(v))

			if indent, ok := pathIndentMap[path]; ok {
				buf.WriteString(strings.Repeat("  ", indent+1))
			} else {
				buf.WriteString(indentStr)
			}

			buf.WriteString(value)

			if indent, ok := pathIndentMap[path]; ok && indentLevel+1 <= indent {
				buf.WriteString("\n")
			} else {
				buf.WriteString(",\n")
			}

			generateCorrectedData(v, pathIndentMap, indentLevel+1)
		}
		buf.WriteString(strings.Repeat("  ", indentLevel))
		buf.WriteString("]")
	default:
		return []byte(getTypeValue(actualData))
	}

	return buf.Bytes()
}



func getTypeValue(value interface{}) string {
	switch value.(type) {
	case map[interface{}]interface{}, []interface{}:
		data, err := yaml.Marshal(value)
		if err != nil {
			return ""
		}
		return "\n" + strings.TrimSpace(string(data))
	default:
		return fmt.Sprintf("%v", value)
	}
}




