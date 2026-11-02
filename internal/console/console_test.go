package console

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
)

func TestDiagnosticStreams(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	require.NoError(t, os.Unsetenv("NO_COLOR"))
	for _, ansi := range []bool{false, true} {
		name := "plain"
		if ansi {
			name = "ansi"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			stdout, err := os.Create(filepath.Join(dir, "stdout"))
			require.NoError(t, err)
			defer stdout.Close()
			stderr, err := os.Create(filepath.Join(dir, "stderr"))
			require.NoError(t, err)
			defer stderr.Close()
			originalStdout, originalStderr := os.Stdout, os.Stderr
			originalColorOutput, originalColorError := color.Output, color.Error
			originalAnsi, originalIsTerminal := isAnsiSequenceSupported, isTerminal
			originalNoColor := color.NoColor
			os.Stdout, os.Stderr = stdout, stderr
			color.Output, color.Error = stdout, stderr
			isAnsiSequenceSupported = ansi
			isTerminal = func(io.Writer) bool { return true }
			color.NoColor = !ansi
			defer func() {
				os.Stdout, os.Stderr = originalStdout, originalStderr
				color.Output, color.Error = originalColorOutput, originalColorError
				isAnsiSequenceSupported, isTerminal = originalAnsi, originalIsTerminal
				color.NoColor = originalNoColor
			}()

			literal, expectedLiteral := " 100%%\n", " 100%\n"
			if ansi {
				literal, expectedLiteral = " 100% %s\n", " 100% %s\n"
			}
			// Function values let us exercise the colored helpers' literal-message
			// behavior without vet treating it as a format error.
			highlight := Highlight
			highlight("result" + literal)
			Highlight("result %s\n", "formatted 100%")
			for _, diagnostic := range []struct {
				name  string
				write func(string, ...interface{})
			}{
				{"updating", func(format string, a ...interface{}) { Fhighlight(os.Stderr, format, a...) }},
				{"reminder", Reminder},
				{"warning", Warn},
				{"error", Error},
				{"success", Success},
			} {
				diagnostic.write(diagnostic.name + literal)
			}

			output, err := os.ReadFile(stdout.Name())
			require.NoError(t, err)
			expected := "result 100%\nresult formatted 100%\n"
			if ansi {
				expected = "\x1b[34mresult 100% %s\n\x1b[0m\x1b[34mresult formatted 100%\n\x1b[0m"
			}
			require.Equal(t, expected, string(output))
			diagnostics, err := os.ReadFile(stderr.Name())
			require.NoError(t, err)
			for _, message := range []string{"updating", "reminder", "warning", "error", "success"} {
				require.Contains(t, string(diagnostics), message+expectedLiteral)
			}
			if ansi {
				for _, code := range []string{"34", "35", "93;40", "31", "32"} {
					require.Contains(t, string(diagnostics), "\x1b["+code+"m")
				}
			} else {
				require.NotContains(t, string(diagnostics), "\x1b[")
			}
		})
	}
}

func TestAnsiDecidedPerWriter(t *testing.T) {
	var term, plain bytes.Buffer
	originalAnsi, originalIsTerminal := isAnsiSequenceSupported, isTerminal
	isAnsiSequenceSupported = true
	isTerminal = func(w io.Writer) bool { return w == &term }
	t.Cleanup(func() { isAnsiSequenceSupported, isTerminal = originalAnsi, originalIsTerminal })

	for _, w := range []io.Writer{&term, &plain} {
		Fprompt(w, "question")
		Fprogress(w, "transferred %d%%", 50)
		ClearPreviousLine(w)
	}
	require.Equal(t, "\x1b[35mquestion\n\x1b[0m\x1b[1Atransferred 50%\x1b[K\n\x1b[1A\x1b[K", term.String())
	require.Equal(t, "questiontransferred 50%\n", plain.String())
}
