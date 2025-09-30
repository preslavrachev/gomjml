package mjml

import (
	"fmt"
	"strings"
)

// ErrorDetail represents a single error detail with line number, message, and tag name.
type ErrorDetail struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
	TagName string `json:"tagName"`
}

// Error addresses errors that occur during the compilation of MJML to HTML.
// It represents errors returned by the MJML engine, providing a general message
// and a list of detailed errors, each including the line number, error message,
// and the tag name where the error occurred.
//
// Error is a direct re-interprtation of the same type present in https://github.com/Boostport/mjml-go
type Error struct {
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details"`
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

// Append merges another Error into this one, combining all details.
// The message from the current error is preserved.
func (e *Error) Append(other *Error) {
	if other == nil {
		return
	}
	e.Details = append(e.Details, other.Details...)
}

func ErrInvalidAttribute(tagName, attrName string, line int) *Error {
	return &Error{
		Message: "MJML compilation error",
		Details: []ErrorDetail{
			{
				Line:    line,
				Message: fmt.Sprintf("Invalid attribute '%s' for tag <%s>", attrName, tagName),
				TagName: tagName,
			},
		},
	}
}
