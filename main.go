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

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
        diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
            A:        difflib.SplitLines(string(data)),
            B:        difflib.SplitLines(string(formattedData)),
            FromFile: "Original",
            ToFile:   "Formatted",
            Context:  3,
        })
        fmt.Printf("Differences in YAML file: %s\n", path)
        fmt.Println(formatDiff(diff))
    }

    return nil
}

func processTerraformFile(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", filename, err)
		return
	}

	ast, diags := hclwrite.ParseConfig(data, filename, nil)
	if diags.HasErrors() {
		fmt.Println("Error parsing file:", filename, diags.Error())
		return
	}

	file := hclwrite.NewEmptyFile()
	file.Body().SetNode(ast.Body())

	formattedData := file.Bytes()

	if string(data) != string(formattedData) {
		fmt.Printf("Proposed changes for %s:\n", filename)
		fmt.Println("-----------------------------------")
		fmt.Println(string(formattedData))
		fmt.Println("-----------------------------------")
	} else {
		fmt.Printf("No changes needed for %s\n", filename)
	}
}

