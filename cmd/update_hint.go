package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/update"
	"github.com/yunxiao-cli/yunxiao/internal/version"
)

// maybeStartUpdateHint kicks an opportunistic Latest Release check.
// Sync with a short timeout when the 24h cache is stale; never fails the user command.
func maybeStartUpdateHint(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	// Skip cobra internals (help is fine to hint on when format allows).
	name := cmd.Name()
	dir, err := config.Dir()
	if err != nil || dir == "" {
		return
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	update.MaybePrintUpdateHint(ctx, update.HintConfig{
		CurrentVersion: version.Version,
		CommandName:    name,
		Format:         globalFormat,
		ConfigDir:      dir,
		Disabled:       update.UpdateCheckDisabled(),
	})
}
