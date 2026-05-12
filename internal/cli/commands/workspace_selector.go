package commands

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func shouldRunDefaultWorkspaceSelector(cmd *cobra.Command, creds *config.CredentialsFile) bool {
	return isTerminal() && len(creds.Workspaces) > 0 && !outputFormatExplicit(cmd)
}

var errWorkspaceSelectionCancelled = errors.New("workspace selection cancelled")

type workspaceSelectionChoice struct {
	Slug          string
	WorkspaceName string
	WorkspaceID   string
	Default       bool
}

func selectDefaultWorkspace(cmd *cobra.Command, creds *config.CredentialsFile) (string, error) {
	choices := workspaceSelectionChoices(creds)
	if len(choices) == 0 {
		return "", errWorkspaceSelectionCancelled
	}

	initial := 0
	for i, choice := range choices {
		if choice.Default {
			initial = i
			break
		}
	}

	input := cmd.InOrStdin()
	if file, ok := input.(*os.File); ok {
		fd := int(file.Fd())
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return "", err
		}
		defer func() {
			_ = term.Restore(fd, oldState)
		}()
	}

	return runWorkspaceSelector(input, cmd.ErrOrStderr(), choices, initial)
}

func workspaceSelectionChoices(creds *config.CredentialsFile) []workspaceSelectionChoice {
	choices := make([]workspaceSelectionChoice, 0, len(creds.Workspaces))
	for _, slug := range creds.SortedWorkspaceSlugs() {
		meta := creds.Workspaces[slug]
		choices = append(choices, workspaceSelectionChoice{
			Slug:          slug,
			WorkspaceName: meta.WorkspaceName,
			WorkspaceID:   meta.WorkspaceID,
			Default:       slug == creds.DefaultWorkspace,
		})
	}
	return choices
}

func runWorkspaceSelector(input io.Reader, out io.Writer, choices []workspaceSelectionChoice, initial int) (string, error) {
	if len(choices) == 0 {
		return "", errWorkspaceSelectionCancelled
	}
	if initial < 0 || initial >= len(choices) {
		initial = 0
	}

	cursor := initial
	selected := initial
	selectedBySpace := false
	reader := bufio.NewReader(input)
	lineCount := renderWorkspaceSelector(out, choices, cursor, selected, 0)

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}

		switch b {
		case 3: // Ctrl-C
			_, _ = fmt.Fprintln(out)
			return "", errWorkspaceSelectionCancelled
		case '\r', '\n':
			if !selectedBySpace {
				selected = cursor
			}
			_, _ = fmt.Fprintln(out)
			return choices[selected].Slug, nil
		case ' ':
			selected = cursor
			selectedBySpace = true
			lineCount = renderWorkspaceSelector(out, choices, cursor, selected, lineCount)
		case 'q', 'Q':
			_, _ = fmt.Fprintln(out)
			return "", errWorkspaceSelectionCancelled
		case 'j', 'J':
			cursor = (cursor + 1) % len(choices)
			lineCount = renderWorkspaceSelector(out, choices, cursor, selected, lineCount)
		case 'k', 'K':
			cursor = (cursor - 1 + len(choices)) % len(choices)
			lineCount = renderWorkspaceSelector(out, choices, cursor, selected, lineCount)
		case 27: // Escape sequence, used by arrow keys.
			next, err := reader.ReadByte()
			if err != nil {
				return "", err
			}
			if next != '[' {
				continue
			}
			code, err := reader.ReadByte()
			if err != nil {
				return "", err
			}
			switch code {
			case 'A':
				cursor = (cursor - 1 + len(choices)) % len(choices)
				lineCount = renderWorkspaceSelector(out, choices, cursor, selected, lineCount)
			case 'B':
				cursor = (cursor + 1) % len(choices)
				lineCount = renderWorkspaceSelector(out, choices, cursor, selected, lineCount)
			}
		}
	}
}

func renderWorkspaceSelector(out io.Writer, choices []workspaceSelectionChoice, cursor, selected, previousLines int) int {
	if previousLines > 0 {
		_, _ = fmt.Fprintf(out, "\x1b[%dA\x1b[J", previousLines)
	}

	lineCount := 0
	writeLine := func(format string, args ...any) {
		_, _ = fmt.Fprintf(out, format+"\r\n", args...)
		lineCount++
	}

	writeLine("Select default Notion workspace")
	writeLine("Use Up/Down or j/k to move, Space to select, Enter to save, q to cancel.")
	writeLine("")

	for i, choice := range choices {
		cursorMarker := " "
		if i == cursor {
			cursorMarker = ">"
		}

		defaultMarker := ""
		if choice.Default {
			defaultMarker = " (current)"
		}

		label := choice.Slug
		if choice.WorkspaceName != "" {
			label = fmt.Sprintf("%s - %s", choice.Slug, choice.WorkspaceName)
		}
		if choice.WorkspaceID != "" {
			label = fmt.Sprintf("%s [%s]", label, shortWorkspaceID(choice.WorkspaceID))
		}

		selectionMarker := " "
		if i == selected {
			selectionMarker = "*"
		}

		writeLine("%s [%s] %s%s", cursorMarker, selectionMarker, label, defaultMarker)
	}
	writeLine("")

	return lineCount
}

func shortWorkspaceID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}
