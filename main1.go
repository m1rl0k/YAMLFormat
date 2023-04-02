package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
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
	if ext != ".tf" {
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	file, diags := hcl.ParseConfig(data, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("Error parsing Terraform file '%s': %s", path, diags.Error())
	}

	hclwriteFile := hclwrite.NewEmptyFile()

	for _, block := range file.Body().Blocks {
		dst := hclwriteFile.Body().AppendNewBlock(block.Type, block.Labels)
		copyAttributes(block, dst)
		copyBlocks(block, dst)
	}

	formattedData := hclwriteFile.Bytes()

	if !hclwrite.Valid(formattedData) {
		return fmt.Errorf("Error formatting Terraform file '%s'", path)
	}

	if !hclwrite.Equal(data, formattedData) {
		fmt.Printf("Differences in Terraform file: %s\n", path)
		diff, _ := hclwrite.HighlightUnifiedDiff(string(data), string(formattedData), path, path, 3)
		fmt.Println(diff)

		if err := ioutil.WriteFile(path, formattedData, 0644); err != nil {
			return err
		}
	}

	return nil
}

func copyAttributes(src *hcl.Block, dst *hclwrite.Block) {
	for _, attr := range src.Body.Attributes {
		dst.SetAttributeValue(attr.Name, attr.Expr)
	}
}

func copyBlocks(src *hcl.Block, dst *hclwrite.Block) {
	for _, block := range src.Body.Blocks {
		dstBlock := dst.Body().AppendNewBlock(block.Type, block.Labels)
		copyAttributes(block, dstBlock)
		copyBlocks(block, dstBlock)
	}
}
