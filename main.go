package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
	} else if ext == ".hcl" {
		return formatHCLFile(path)
	} else if ext == ".tf" {
		return formatTerraformFile(path)
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
		if err := ioutil.WriteFile(path, formattedData, 0644); err != nil {
			return err
		}
		fmt.Printf("Formatted YAML file: %s\n", path)
	}

	return nil
}

func formatHCLFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	file, diag := hclsyntax.ParseConfig(data, path, hcl.Pos{Line: 1, Column: 1})
	if diag.HasErrors() {
		return diag
	}

	formattedData := hclwrite.NewEmptyFile()
	_ = hclwrite.MergeFiles([]*hclwrite.File{formattedData, hclwrite.NewFileFromTokens(file)})
	formattedBytes := formattedData.Bytes()

	if !bytes.Equal(data, formattedBytes) {
		if err := ioutil.WriteFile(path, formattedBytes, 0644); err != nil {
			return err
		}
		fmt.Printf("Formatted HCL file: %s\n", path)
	}

	return nil
}

func formatTerraformFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	file, diag := hclsyntax.ParseConfig(data, path, hcl.Pos{Line: 1, Column: 1})
	if diag.HasErrors() {
		return diag
	}

	formattedData := hclwrite.NewEmptyFile()
	_ = hclwrite.MergeFiles([]*hclwrite.File{formattedData, hclwrite.NewFileFromTokens(file)})
	formattedBytes := formattedData.Bytes()

	if !bytes.Equal(data, formattedBytes) {
		if err := ioutil.WriteFile(path, formattedBytes, 0644); err != nil {
			return err
		}
		fmt.Printf("Formatted Terraform file: %s\n", path)
	}

	return nil
}
