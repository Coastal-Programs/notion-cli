package commands

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/config"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterDoctorCommand registers the doctor command.
func RegisterDoctorCommand(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"diagnose", "healthcheck"},
		Short:   "Run diagnostics",
		Long:    "Run comprehensive health checks on your Notion CLI setup.",
		RunE:    runDoctor,
	}
	addOutputFlags(cmd)
	root.AddCommand(cmd)
}

type checkResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "pass", "fail", "warn"
	Message string `json:"message"`
}

func runDoctor(cmd *cobra.Command, args []string) error {
	start := time.Now()
	var checks []checkResult

	// Check 1: Go version.
	goVer := runtime.Version()
	checks = append(checks, checkResult{
		Name:    "Go Runtime",
		Status:  "pass",
		Message: goVer,
	})

	// Check 2: Token configured.
	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		cfg, err := config.LoadConfig()
		if err == nil && cfg.Token != "" {
			token = cfg.Token
		}
	}
	if token == "" {
		checks = append(checks, checkResult{
			Name:    "API Token",
			Status:  "fail",
			Message: "NOTION_TOKEN not set. Run 'notion-cli config set-token <token>' or set NOTION_TOKEN env var.",
		})
	} else {
		// Mask token for display.
		masked := token
		if len(token) > 10 {
			masked = token[:7] + "***..." + token[len(token)-3:]
		}
		checks = append(checks, checkResult{
			Name:    "API Token",
			Status:  "pass",
			Message: fmt.Sprintf("Configured (%s)", masked),
		})
	}

	// Check 3: Token format.
	if token != "" {
		if strings.HasPrefix(token, "secret_") || strings.HasPrefix(token, "ntn_") {
			checks = append(checks, checkResult{
				Name:    "Token Format",
				Status:  "pass",
				Message: "Valid prefix detected",
			})
		} else {
			checks = append(checks, checkResult{
				Name:    "Token Format",
				Status:  "warn",
				Message: "Token should start with 'secret_' or 'ntn_'",
			})
		}
	}

	// Check 4: Network connectivity.
	conn, err := net.DialTimeout("tcp", "api.notion.com:443", 5*time.Second)
	if err != nil {
		checks = append(checks, checkResult{
			Name:    "Network",
			Status:  "fail",
			Message: fmt.Sprintf("Cannot reach api.notion.com:443: %s", err),
		})
	} else {
		conn.Close()
		checks = append(checks, checkResult{
			Name:    "Network",
			Status:  "pass",
			Message: "api.notion.com:443 reachable",
		})
	}

	// Check 5: API connection (only if token is set).
	if token != "" {
		client, err := newClient()
		if err == nil {
			apiStart := time.Now()
			_, apiErr := client.UsersMe(cmd.Context())
			latency := time.Since(apiStart)
			if apiErr != nil {
				checks = append(checks, checkResult{
					Name:    "API Connection",
					Status:  "fail",
					Message: fmt.Sprintf("API call failed: %s", apiErr),
				})
			} else {
				checks = append(checks, checkResult{
					Name:    "API Connection",
					Status:  "pass",
					Message: fmt.Sprintf("Connected (latency: %dms)", latency.Milliseconds()),
				})
			}
		}
	}

	// Check 6: Config directory.
	dataDir := config.GetDataDir()
	if dataDir != "" {
		if info, err := os.Stat(dataDir); err == nil && info.IsDir() {
			checks = append(checks, checkResult{
				Name:    "Data Directory",
				Status:  "pass",
				Message: dataDir,
			})
		} else {
			checks = append(checks, checkResult{
				Name:    "Data Directory",
				Status:  "warn",
				Message: fmt.Sprintf("%s does not exist (will be created on first use)", dataDir),
			})
		}
	}

	// Check 7: Workspace cache.
	cacheFile := ""
	if dataDir != "" {
		cacheFile = dataDir + "/databases.json"
	}
	if cacheFile != "" {
		if info, err := os.Stat(cacheFile); err == nil {
			age := time.Since(info.ModTime())
			if age > 24*time.Hour {
				checks = append(checks, checkResult{
					Name:    "Workspace Cache",
					Status:  "warn",
					Message: fmt.Sprintf("Stale (%.0f hours old). Run 'notion-cli sync' to refresh.", age.Hours()),
				})
			} else {
				checks = append(checks, checkResult{
					Name:    "Workspace Cache",
					Status:  "pass",
					Message: fmt.Sprintf("Fresh (%.0f minutes old)", age.Minutes()),
				})
			}
		} else {
			checks = append(checks, checkResult{
				Name:    "Workspace Cache",
				Status:  "warn",
				Message: "Not synced yet. Run 'notion-cli sync' to cache workspace databases.",
			})
		}
	}

	// Build summary.
	passCount := 0
	failCount := 0
	warnCount := 0
	for _, c := range checks {
		switch c.Status {
		case "pass":
			passCount++
		case "fail":
			failCount++
		case "warn":
			warnCount++
		}
	}

	data := map[string]any{
		"checks":  checks,
		"summary": fmt.Sprintf("%d passed, %d warnings, %d failed", passCount, warnCount, failCount),
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "doctor", start)
	return nil
}
