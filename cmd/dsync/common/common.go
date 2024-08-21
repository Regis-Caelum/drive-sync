package common

import (
	"os"
	"strings"
)

func PathExist(absPath string) bool {
	_, err := os.Stat(absPath)
	return !os.IsNotExist(err)
}

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
