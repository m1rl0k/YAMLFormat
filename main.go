package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			FromFile: "Original",
			ToFile:   "Formatted",
			Context:  3,
		})
		fmt.Printf("Differences in YAML file: %s\n", path)
		fmt.Println(diff)

		if err := ioutil.WriteFile(path, formattedData, 0644); err != nil {
			return err
		}
	}

	return nil
}

func formatTerraformFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	file, diags := hclsyntax.ParseConfig(data, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("Error parsing Terraform file '%s': %s", path, diags.Error())
	}

	hclwriteFile := hclwrite.NewEmptyFile()
	body := hclwriteFile.Body()
	file.Body().Walk(func(node hclsyntax.Node) (bool, error) {
		switch n := node.(type) {
		case *hclsyntax.Attribute:
			attr := body.SetAttribute(n.Name, hclwrite.ExpressionFromExpr(n.Expr))
			attr.Range = n.Range()
			attr.SrcRange = n.SrcRange
			attr.Expr.Range = n.Expr.Range()
			attr.Expr.SrcRange = n.Expr.SrcRange()
		case *hclsyntax.Block:
			block := body.AppendNewBlock(n.Type, n.Labels)
			block.Range = n.Range()
			block.SrcRange = n.SrcRange
			n.Body.Walk(func(node hclsyntax.Node) (bool, error) {
				block.Body().AppendUnstructuredTokens(hclwrite.TokensForNode(node))
				return true, nil
			})
		}
		return true, nil
	})

	formattedData := hclwrite.Format(hclwriteFile.Bytes())

	if !strings.EqualFold(string(data), string(formattedData)) {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			FromFile: "Original",
			ToFile:   "Formatted",
			Context:  3,
		})
		fmt.Printf("Differences in Terraform file: %s\n", path)
		fmt.Println(diff)

		if err := ioutil.WriteFile(path, formattedData, 0644); err != nil {
			return err
		}
	}

	return nil
}
