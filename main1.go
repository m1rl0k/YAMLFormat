package tf


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

    config := hclsyntax.EncodeConfig(file.Body, &hclwrite.Indent{})
    return config.Bytes(), nil
}
