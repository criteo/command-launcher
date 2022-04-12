package helper

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// A contextual help error not only provides the error message, but also a suggestion
// for user to correct them in the next command
//
// for example, when command fails to find necessary files in the workspace,
// an actionable error will be helpful to remind the user double check the workspace,
// or specify a different one with dedicated flag
func ContextualHelpError(err error, suggestions ...string) error {
	suggestionsText := strings.Join(suggestions, "\n")
	return fmt.Errorf("%s\n%s", err.Error(), color.MagentaString(suggestionsText))
}
