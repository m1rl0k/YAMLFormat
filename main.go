func formatHCLFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	file, diags := hclsyntax.ParseConfig(data, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("Error parsing HCL file '%s': %s", path, diags.Error())
	}

	hclwriteFile := hclwrite.NewEmptyFile()
	if err := hclwriteFile.Body().Build(file.Body()); err != nil {
		return err
	}

	formattedData := hclwriteFile.Bytes()

	if !strings.EqualFold(string(data), string(formattedData)) {
		if err := ioutil.WriteFile(path, formattedData, 0644); err != nil {
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

	file, diags := hclsyntax.ParseConfig(data, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("Error parsing Terraform file '%s': %s", path, diags.Error())
	}

	jsonBytes, err := json.Marshal(file.Body())

	if err != nil {
		return err
	}

	hclwriteFile, err := hclwrite.ParseJSON(jsonBytes, path)
	if err != nil {
		return err
	}

	formattedData := hclwrite.Format(hclwriteFile.Bytes())

	if !strings.EqualFold(string(data), string(formattedData)) {
		if err := ioutil.WriteFile(path, formattedData, 0644); err != nil {
			return err
		}
		fmt.Printf("Formatted Terraform file: %s\n", path)
	}

	return nil
}

