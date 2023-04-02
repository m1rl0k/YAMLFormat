package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v3"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := filepath.Walk(dir, processFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	ext := filepath.Ext(path)
	if ext == ".yaml" || ext == ".yml" {
		return formatYAMLFile(path)
	}

	return nil
}

func formatYAMLFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	formattedData, err := formatYAML(data)
	if err != nil {
		return err
	}

	if !strings.EqualFold(string(data), string(formattedData)) {
		fmt.Printf("Differences in YAML file: %s\n", path)
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			FromFile: "Original",
			ToFile:   "Formatted",
			Context:  3,
		})
		fmt.Println(diff)
	} else {
		fmt.Printf("No differences in YAML file: %s\n", path)
	}

	fmt.Println(strings.Repeat("-", 80))

	return nil
}

func formatYAML(data []byte) ([]byte, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, err
	}

	formattedData, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, err
	}

	return formattedData, nil
}
