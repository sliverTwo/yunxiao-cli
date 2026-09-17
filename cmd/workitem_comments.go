package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var workitemCommentsCmd = &cobra.Command{Use: "comments", Short: "Work item comments"}

var workitemCommentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List comments on a work item",
	Long:  "Risk: read\nHTTP: GET .../workitems/{id}/comments\n\nDefault order: newest first (by gmtModified/gmtCreate). Use --sort asc for oldest first.",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		id, _ := cmd.Flags().GetString("id")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		sortFlag, _ := cmd.Flags().GetString("sort")
		if err := requireFlags("id", id); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/comments")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"page": strconv.Itoa(page), "perPage": strconv.Itoa(perPage)}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, afterSortByTime(sortFlag, nil)))
	},
}

var workitemCommentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Add a comment to a work item",
	Long:  "Risk: write\nHTTP: POST .../workitems/{id}/comments",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		id, _ := cmd.Flags().GetString("id")
		content, _ := cmd.Flags().GetString("content")
		if err := requireFlags("id", id, "content", content); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/comments")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"content": content}
		handleErr(runJSONMutating(cmd.Context(), c, "workitem comment", risk.Write, "POST", path, nil, body, nil))
	},
}
