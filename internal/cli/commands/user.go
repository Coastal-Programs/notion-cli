package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/notion"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterUserCommands registers all user subcommands under root.
func RegisterUserCommands(root *cobra.Command) {
	userCmd := &cobra.Command{
		Use:     "user",
		Aliases: []string{"u"},
		Short:   "User operations",
		Long:    "List, retrieve, and inspect Notion workspace users.",
	}

	userCmd.AddCommand(
		newUserListCmd(),
		newUserRetrieveCmd(),
		newUserBotCmd(),
	)

	root.AddCommand(userCmd)
}

// --- user list ---

func newUserListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List all users",
		Long:    "List all users in the Notion workspace.",
		Args:    cobra.NoArgs,
		RunE:    runUserList,
	}
	addOutputFlags(cmd)
	return cmd
}

func runUserList(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	var allUsers []any
	query := notion.QueryParams{}
	pageCount := 0

	for {
		pageCount++
		if pageCount > maxPaginationPages {
			fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
			break
		}

		result, err := client.UsersList(cmd.Context(), query)
		if err != nil {
			return handleError(cmd, err)
		}

		users, _ := result["results"].([]any)
		allUsers = append(allUsers, users...)

		hasMore, _ := result["has_more"].(bool)
		nextCursor, _ := result["next_cursor"].(string)

		if !hasMore || nextCursor == "" {
			break
		}
		query.StartCursor = nextCursor
	}

	// Build table-friendly output
	var tableData []map[string]any
	for _, u := range allUsers {
		user, ok := u.(map[string]any)
		if !ok {
			continue
		}
		row := map[string]any{
			"id":   user["id"],
			"name": user["name"],
			"type": user["type"],
		}
		if person, ok := user["person"].(map[string]any); ok {
			row["email"] = person["email"]
		}
		tableData = append(tableData, row)
	}

	data := map[string]any{
		"results":      tableData,
		"result_count": len(tableData),
	}

	p.PrintSuccess(data, "user list", start)
	return nil
}

// --- user retrieve ---

func newUserRetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retrieve <user_id>",
		Aliases: []string{"r", "get"},
		Short:   "Retrieve a user",
		Long:    "Retrieve a Notion user by ID.",
		Args:    cobra.ExactArgs(1),
		RunE:    runUserRetrieve,
	}
	addOutputFlags(cmd)
	return cmd
}

func runUserRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	userID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.UserRetrieve(cmd.Context(), userID)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "user retrieve", start)
	return nil
}

// --- user bot ---

func newUserBotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bot",
		Aliases: []string{"me", "b"},
		Short:   "Retrieve bot user",
		Long:    "Retrieve the bot user associated with the current API token.",
		Args:    cobra.NoArgs,
		RunE:    runUserBot,
	}
	addOutputFlags(cmd)
	return cmd
}

func runUserBot(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.UsersMe(cmd.Context())
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "user bot", start)
	return nil
}
