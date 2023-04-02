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
	if err := filepath.Walk(".", processFile); err != nil {
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

	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return err
	}

	formattedData, err := yaml.Marshal(yamlData)
	if err != nil {
		return err
	}

	if !strings.EqualFold(string(data), string(formattedData)) {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			FromFile: "Original",
			ToFile:   "Formatted",
			Context:  3,
		})
		fmt.Printf("Differences in YAML file: %s\n", path)
		fmt.Println(diff)
	}

	return nil
}
