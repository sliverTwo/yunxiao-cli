package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var workitemSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search work items",
	Long:  "Risk: read\nHTTP: POST .../workitems:search",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		category, _ := cmd.Flags().GetString("category")
		spaceID, _ := cmd.Flags().GetString("space-id")
		spaceID, err := resolveSpaceIDFlag(spaceID)
		if err != nil {
			handleErr(err)
			return
		}
		assignedTo, _ := cmd.Flags().GetString("assigned-to")
		creator, _ := cmd.Flags().GetString("creator")
		subject, _ := cmd.Flags().GetString("subject")
		statusStage, _ := cmd.Flags().GetString("status-stage")
		workitemType, _ := cmd.Flags().GetString("type")
		priority, _ := cmd.Flags().GetString("priority")
		orderBy, _ := cmd.Flags().GetString("order-by")
		sort, _ := cmd.Flags().GetString("sort")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		assignedTo, err = resolveSelfID(cmd.Context(), c, assignedTo)
		if err != nil {
			handleErr(err)
			return
		}
		creator, err = resolveSelfID(cmd.Context(), c, creator)
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems:search")
		if err != nil {
			handleErr(err)
			return
		}
		var filters []any
		filters = appendUserFilter(filters, "assignedTo", assignedTo)
		filters = appendUserFilter(filters, "creator", creator)
		if subject != "" {
			filters = append(filters, map[string]any{
				"className": "string", "fieldIdentifier": "subject", "format": "input",
				"operator": "CONTAINS", "value": []string{subject},
			})
		}
		if statusStage != "" {
			filters = append(filters, map[string]any{
				"className": "statusStage", "fieldIdentifier": "statusStage", "format": "list",
				"operator": "CONTAINS", "value": splitCSV(statusStage),
			})
		}
		if workitemType != "" {
			filters = append(filters, map[string]any{
				"className": "workitemType", "fieldIdentifier": "workitemType", "format": "list",
				"operator": "CONTAINS", "value": splitCSV(workitemType),
			})
		}
		if priority != "" {
			filters = append(filters, map[string]any{
				"className": "option", "fieldIdentifier": "priority", "format": "list",
				"operator": "CONTAINS", "value": splitCSV(priority),
			})
		}
		body := map[string]any{
			"category": category,
			"orderBy":  orderBy,
			"sort":     sort,
			"page":     page,
			"perPage":  perPage,
		}
		body["spaceId"] = spaceID
		if len(filters) > 0 {
			conds := map[string]any{"conditionGroups": []any{filters}}
			cb, _ := json.Marshal(conds)
			body["conditions"] = string(cb)
		}
		handleErr(runRead(cmd.Context(), c, "POST", path, nil, body, map[string]any{"risk": risk.Read}, nil))
	},
}
