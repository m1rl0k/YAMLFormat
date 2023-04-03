func showDiff(path, original, formatted string) {
    dmp := diffmatchpatch.New()
    diffs := dmp.DiffMain(original, formatted, false)

    if len(diffs) > 1 {
        fmt.Printf("Differences in file %s:\n", path)

        // convert diffs to deltas
        deltas := dmp.DiffToDelta(diffs)

        // loop through deltas
        pos := 0
        for _, delta := range deltas {
            deltaType := delta[0]

            switch deltaType {
            case '+':
                fmt.Println("Added:")
            case '-':
                fmt.Println("Removed:")
            case '=':
                length, _ := strconv.Atoi(string(delta[1:]))
                pos += length
                continue
            }

            start := pos - 12
            if start < 0 {
                start = 0
            }
            end := pos + 12
            if end >= len(original) {
                end = len(original) - 1
            }

            fmt.Println("Original:")
            for i := start; i <= end; i++ {
                fmt.Printf("%4d: %s\n", i+1, string(original[i]))
            }

            fmt.Println("Formatted:")
            for i := start; i <= end; i++ {
                fmt.Printf("%4d: %s\n", i+1, string(formatted[i]))
            }

            length, _ := strconv.Atoi(string(delta[1:]))
            pos += length
        }
    }
}


func suggestFixForLine(line string) string {
	fixedLine := line

	// Check for missing colons
	if !strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
		index := strings.IndexFunc(line, unicode.IsLetter)
		if index != -1 {
			fixedLine = line[:index+1] + ": " + line[index+1:]
		}
	}

	// Check for missing spaces after colons
	colonIndex := strings.Index(line, ":")
	if colonIndex != -1 && len(line) > colonIndex+1 && line[colonIndex+1] != ' ' {
		fixedLine = line[:colonIndex+1] + " " + line[colonIndex+1:]
	}

	// Fix incorrect indentation
	indentation := 0
	for _, char := range line {
		if char == ' ' {
			indentation++
		} else {
			break
		}
	}

	correctIndentation := (indentation / 2) * 2
	if correctIndentation != indentation {
		fixedLine = strings.Repeat(" ", correctIndentation) + strings.TrimSpace(line)
	}

	return fixedLine
}
