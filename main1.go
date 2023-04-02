package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/hashicorp/hcl/v2/hclwrite"

    "github.com/pmezard/go-difflib/difflib"
    "github.com/hashicorp/hcl/v2/hclsyntax"
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
        if ext == ".tf" || ext == ".tfvars" {
            return formatTerraformFile(path)
        }

        return nil
    }); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func formatTerraformFile(path string) error {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return err
    }

    formattedData, err := formatTerraform(data)
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
        fmt.Printf("Differences in Terraform file: %s\n", path)
        fmt.Println(formatDiff(diff))
    }

    return nil
}

func formatTerraform(data []byte) ([]byte, error) {
    parser := hclsyntax.NewParser()
    file, diags := parser.ParseHCL(data, "input.tf")
    if diags.HasErrors() {
        return nil, fmt.Errorf("failed to parse HCL: %v", diags)
    }

    config := hclsyntax.EncodeConfig(file.Body, &hclwrite.TabIndent{})
    return config.Bytes(), nil
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
