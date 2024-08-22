package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FormatSection(header string, content string) string {
	out := ""

	// Add section header
	if header != "" {
		out += header + ":\n"
	}

	// Indent the content
	for _, line := range strings.Split(content, "\n") {
		if line != "" {
			out += "  "
		}

		out += line + "\n"
	}

	if header != "" {
		// Section separator (when rendering a full section
		out += "\n"
	} else {
		// Remove last newline when rendering partial section
		out = strings.TrimSuffix(out, "\n")
	}

	return out
}

func PrintTable(headers []string, rows [][]string) {
	// Calculate the width of each column
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}
	for _, row := range rows {
		for i, col := range row {
			if len(col) > colWidths[i] {
				colWidths[i] = len(col)
			}
		}
	}

	// Create top, middle, and bottom separator lines with curves
	topSeparator := "╭"
	midSeparator := "├"
	bottomSeparator := "╰"
	for i, width := range colWidths {
		if i > 0 {
			topSeparator += "┬"
			midSeparator += "┼"
			bottomSeparator += "┴"
		}
		topSeparator += strings.Repeat("─", width+2)
		midSeparator += strings.Repeat("─", width+2)
		bottomSeparator += strings.Repeat("─", width+2)
	}
	topSeparator += "╮"
	midSeparator += "┤"
	bottomSeparator += "╯"

	// Print the headers
	fmt.Println(topSeparator)
	fmt.Print("│")
	for i, header := range headers {
		fmt.Printf(" %-*s │", colWidths[i], header)
	}
	fmt.Println()
	fmt.Println(midSeparator)

	// Print the rows
	for _, row := range rows {
		fmt.Print("│")
		for i, col := range row {
			fmt.Printf(" %-*s │", colWidths[i], col)
		}
		fmt.Println()
	}
	fmt.Println(bottomSeparator)
}

func PathExist(absPath string) bool {
	_, err := os.Stat(absPath)
	return !os.IsNotExist(err)
}

func IsDir(absPath string) bool {
	info, err := os.Stat(absPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func IsHiddenPath(path string) bool {
	// Check if the path or any segment of the path starts with a dot
	segments := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, segment := range segments {
		if strings.HasPrefix(segment, ".") {
			return true
		}
	}
	return false
}
