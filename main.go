
 package main

 import (
        "bytes"
 	"encoding/json"
 	"fmt"
 	"io/ioutil"
 	"os"
 	"path/filepath"
 	"strings"

 	"github.com/hashicorp/hcl/v2"
 	"github.com/hashicorp/hcl/v2/hclwrite"
 	"gopkg.in/yaml.v3"
 	"github.com/hashicorp/hcl/v2/hclsyntax"
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
		return fmt.Errorf("failed to read file: %w", err)
	}

	file, diags := hclsyntax.ParseConfig(data, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("failed to parse HCL file '%s': %w", path, diags)
	}

	hclwriteFile := hclwrite.NewEmptyFile()
	if err := file.Build(hclwriteFile); err != nil {
		return fmt.Errorf("failed to build HCL file: %w", err)
	}

	formattedData := hclwriteFile.Bytes()

	if !bytes.Equal(data, formattedData) {
		if err := ioutil.WriteFile(path, formattedData, 0644); err != nil {
			return fmt.Errorf("failed to write formatted HCL file: %w", err)
		}
		fmt.Printf("Formatted HCL file: %s\n", path)
	}

	return nil
}

func formatTerraformFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	file, diags := hclsyntax.ParseConfig(data, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("failed to parse Terraform file '%s': %w", path, diags)
	}

	jsonBytes, err := json.Marshal(file.Body())
	if err != nil {
		return fmt.Errorf("failed to convert HCL to JSON: %w", err)
	}

	hclwriteFile, err := hclwrite.ParseJSON(jsonBytes, path)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	formattedData := hclwrite.Format(hclwriteFile.Bytes())

	if !bytes.Equal(data, formattedData) {
		if err := ioutil.WriteFile(path, formattedData, 0644); err != nil {
			return fmt.Errorf("failed to write formatted Terraform file: %w", err)
		}
		fmt.Printf("Formatted Terraform file: %s\n", path)
	}

	return nil
}


