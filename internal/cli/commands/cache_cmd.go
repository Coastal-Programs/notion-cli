package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/config"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterCacheCommands registers cache subcommands.
func RegisterCacheCommands(root *cobra.Command) {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Cache management",
		Long:  "View cache status and statistics.",
	}

	cacheCmd.AddCommand(newCacheInfoCmd())
	root.AddCommand(cacheCmd)
}

func newCacheInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info",
		Aliases: []string{"stats", "status"},
		Short:   "Show cache status",
		Long:    "Show cache statistics including workspace cache and disk cache status.",
		RunE:    runCacheInfo,
	}
	addOutputFlags(cmd)
	return cmd
}

func runCacheInfo(cmd *cobra.Command, args []string) error {
	start := time.Now()

	cfg, err := config.LoadConfig()
	if err != nil {
		return handleError(cmd, fmt.Errorf("load config: %w", err))
	}

	data := map[string]any{
		"cache_enabled":      cfg.CacheEnabled,
		"cache_max_size":     cfg.CacheMaxSize,
		"disk_cache_enabled": cfg.DiskCacheEnabled,
	}

	// Check workspace cache.
	dataDir := config.GetDataDir()
	if dataDir != "" {
		cacheFile := filepath.Join(dataDir, "databases.json")
		if info, err := os.Stat(cacheFile); err == nil {
			data["workspace_cache"] = map[string]any{
				"path":     cacheFile,
				"size":     info.Size(),
				"modified": info.ModTime().Format(time.RFC3339),
				"age":      fmt.Sprintf("%.0f minutes", time.Since(info.ModTime()).Minutes()),
			}
		} else {
			data["workspace_cache"] = map[string]any{
				"status": "not synced",
			}
		}

		// Check disk cache directory.
		diskCacheDir := filepath.Join(dataDir, "cache")
		if info, err := os.Stat(diskCacheDir); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(diskCacheDir)
			var totalSize int64
			for _, e := range entries {
				if fi, err := e.Info(); err == nil {
					totalSize += fi.Size()
				}
			}
			data["disk_cache"] = map[string]any{
				"path":       diskCacheDir,
				"entries":    len(entries),
				"total_size": totalSize,
			}
		} else {
			data["disk_cache"] = map[string]any{
				"status": "not initialized",
			}
		}
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "cache info", start)
	return nil
}
