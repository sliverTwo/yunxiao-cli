package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/client"
)

var workitemCmd = &cobra.Command{
	Use:     "workitem",
	Aliases: []string{"wi"},
	Short:   "Projex work items: search, get, comments, relations",
	Long: `Work item typed commands.

  yunxiao workitem search --assigned-to self --category Req
  yunxiao workitem get --id <id|ZYPT-xxxx>   # or positional: workitem get ZYPT-xxxx
  yunxiao workitem +transition --id <id|serial> --to <alias|statusId> --dry-run
  yunxiao workitem +bug-transition --id ZYPT-xxxx --to processing --dry-run
  yunxiao workitem +bug-create --title "…" --description "…" --expected-completion YYYY-MM-DD --sprint <id> --dry-run
  yunxiao workitem +explore-workflow --type-id <id> --cleanup --dry-run
  yunxiao workitem comments list --id <id>
  yunxiao workitem comment --id <id> --content '…' [--dry-run]
  yunxiao workitem create|update|delete
  yunxiao workitem relations list|create|delete
  yunxiao workitem attachments list|create
  yunxiao workitem types list --space-id <id> --category Req

Risk: search/get/comments/types/relations.list/attachments.list=read; +transition/+bug-transition/+bug-create/+explore-workflow/create/comment/update/relations/attachments.create=write; delete=high-risk-write`,
}

func resolveSelfID(ctx context.Context, c *client.Client, v string) (string, error) {
	if v != "self" {
		return v, nil
	}
	var user map[string]any
	if err := c.Get(ctx, "/oapi/v1/platform/user", nil, &user); err != nil {
		return "", err
	}
	id, _ := user["id"].(string)
	return id, nil
}

func appendUserFilter(filters []any, field, userID string) []any {
	if userID == "" {
		return filters
	}
	return append(filters, map[string]any{
		"className": "user", "fieldIdentifier": field, "format": "list",
		"operator": "CONTAINS", "value": []string{userID},
	})
}
