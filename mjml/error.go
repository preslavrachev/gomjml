package mjml

import (
	"fmt"
	"strings"
)

// Error addresses errors that occur during the compilation of MJML to HTML.
// It represents errors returned by the MJML engine, providing a general message
// and a list of detailed errors, each including the line number, error message,
// and the tag name where the error occurred.
//
// Error is a direct re-interprtation of the same type present in https://github.com/Boostport/mjml-go
type Error struct {
	Message string `json:"message"`
	Details []struct {
		Line    int    `json:"line"`
		Message string `json:"message"`
		TagName string `json:"tagName"`
	} `json:"details"`
}

func (e Error) Error() string {
	var sb strings.Builder

	sb.WriteString(e.Message)

	numDetails := len(e.Details)

	if numDetails > 0 {
		sb.WriteString(":\n")
	}

	for i, detail := range e.Details {
		sb.WriteString(fmt.Sprintf("- Line %d of (%s) - %s", detail.Line, detail.TagName, detail.Message))

		if i != numDetails-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func ErrInvalidAttribute(tagName, attrName string, line int) *Error {
	return &Error{
		Message: "MJML compilation error",
		Details: []struct {
			Line    int    `json:"line"`
			Message string `json:"message"`
			TagName string `json:"tagName"`
		}{
			{
				Line:    line,
				Message: fmt.Sprintf("Invalid attribute '%s' for tag <%s>", attrName, tagName),
				TagName: tagName,
			},
		},
	}
}
