package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
	"github.com/yunxiao-cli/yunxiao/internal/zhiyi"
)

var workitemRelationsCmd = &cobra.Command{Use: "relations", Short: "Work item relation records"}

var workitemRelationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List relation records for a work item",
	Long: `Risk: read
HTTP: GET .../workitems/{id}/relationRecords?relationType=...

Relation types seen working in sandbox sims: ASSOCIATED, DEPEND_ON.
RELATED and PARENT_SUB often fail type constraints; for Task hierarchy prefer --parent-id on workitem create.

After listing, each record's resourceId is resolved via GET workitem (bounded concurrency)
and enriched best-effort with serial_number, subject, category/resourceType, and url
(using profile/item space). On resolve failure the original record is kept with resolve_error.

  yunxiao workitem relations list --id <id> --relation-type ASSOCIATED`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		id, _ := cmd.Flags().GetString("id")
		relType, _ := cmd.Flags().GetString("relation-type")
		if err := requireFlags("id", id, "relation-type", relType); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/relationRecords")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"relationType": relType}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, func(out any, meta map[string]any) (any, map[string]any) {
			space := profileSpaceID()
			out = zhiyi.EnrichRelationRecords(out, func(rid string) (map[string]any, error) {
				return fetchWorkItemMap(cmd.Context(), c, rid)
			}, space, 6)
			meta["enriched"] = true
			return out, meta
		}))
	},
}

var workitemRelationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a relation between work items (write)",
	Long: `Risk: write
HTTP: POST .../workitems/{id}/relationRecords

Working relation-type values (sandbox): ASSOCIATED, DEPEND_ON.
RELATED / PARENT_SUB often return 不满足关系对应的类型约束.
Task→parent: use workitem create --parent-id instead of PARENT_SUB.

  yunxiao workitem relations create --id <id> --related-id <rid> --relation-type ASSOCIATED --dry-run`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		id, _ := cmd.Flags().GetString("id")
		related, _ := cmd.Flags().GetString("related-id")
		relType, _ := cmd.Flags().GetString("relation-type")
		if err := requireFlags("id", id, "related-id", related, "relation-type", relType); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/relationRecords")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"relationType": relType, "workitemId": related}
		handleErr(runJSONMutating(cmd.Context(), c, "workitem relations create", risk.Write, "POST", path, nil, body, nil))
	},
}

var workitemRelationsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a relation between work items (write)",
	Long: `Risk: write
HTTP: DELETE .../workitems/{id}/relationRecords (JSON body)

Use the same --relation-type as create (e.g. ASSOCIATED / DEPEND_ON).`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		id, _ := cmd.Flags().GetString("id")
		related, _ := cmd.Flags().GetString("related-id")
		relType, _ := cmd.Flags().GetString("relation-type")
		if err := requireFlags("id", id, "related-id", related, "relation-type", relType); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/relationRecords")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"relationType": relType, "workitemId": related}
		handleErr(runJSONMutating(cmd.Context(), c, "workitem relations delete", risk.Write, "DELETE", path, nil, body, nil))
	},
}
