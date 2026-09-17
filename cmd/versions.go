package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

// Command name is "versions" (plural) to avoid clashing with root --version.
var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Projex project versions",
	Long: `Version typed commands (operations/projex/version.ts).

  yunxiao versions list --space-id <projectId>
  yunxiao versions create|update|delete

Risk: list=read; create/update=write; delete=high-risk-write`,
}

var versionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List versions in a project",
	Long:  "Risk: read\nHTTP: GET .../projects/{space}/versions",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		spaceID, _ := cmd.Flags().GetString("space-id")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		if err := requireFlags("space-id", spaceID); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/projects/"+spaceID+"/versions")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{}
		if page > 0 {
			q["page"] = strconv.Itoa(page)
		}
		if perPage > 0 {
			q["perPage"] = strconv.Itoa(perPage)
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var versionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a project version (write)",
	Long:  "Risk: write\nHTTP: POST .../projects/{space}/versions",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		spaceID, _ := cmd.Flags().GetString("space-id")
		name, _ := cmd.Flags().GetString("name")
		owners, _ := cmd.Flags().GetString("owners")
		startDate, _ := cmd.Flags().GetString("start-date")
		publishDate, _ := cmd.Flags().GetString("publish-date")
		if err := requireFlags("space-id", spaceID, "name", name, "owners", owners); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/projects/"+spaceID+"/versions")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"name": name, "owners": splitCSV(owners)}
		if startDate != "" {
			body["startDate"] = startDate
		}
		if publishDate != "" {
			body["publishDate"] = publishDate
		}
		handleErr(runJSONMutating(cmd.Context(), c, "versions create", risk.Write, "POST", path, nil, body, nil))
	},
}

var versionsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a project version (write)",
	Long:  "Risk: write\nHTTP: PUT .../projects/{space}/versions/{id}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		spaceID, _ := cmd.Flags().GetString("space-id")
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		owners, _ := cmd.Flags().GetString("owners")
		startDate, _ := cmd.Flags().GetString("start-date")
		publishDate, _ := cmd.Flags().GetString("publish-date")
		if err := requireFlags("space-id", spaceID, "id", id, "name", name); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/projects/"+spaceID+"/versions/"+id)
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"name": name}
		if owners != "" {
			body["owners"] = splitCSV(owners)
		}
		if startDate != "" {
			body["startDate"] = startDate
		}
		if publishDate != "" {
			body["publishDate"] = publishDate
		}
		handleErr(runJSONMutating(cmd.Context(), c, "versions update", risk.Write, "PUT", path, nil, body, nil))
	},
}

var versionsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a project version (high-risk-write)",
	Long:  "Risk: high-risk-write\nHTTP: DELETE .../projects/{space}/versions/{id}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		spaceID, _ := cmd.Flags().GetString("space-id")
		id, _ := cmd.Flags().GetString("id")
		if err := requireFlags("space-id", spaceID, "id", id); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/projects/"+spaceID+"/versions/"+id)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, "versions delete", risk.HighRiskWrite, "DELETE", path, nil, nil, nil))
	},
}

func init() {
	versionsListCmd.Flags().String("space-id", "", "project/space id (required)")
	versionsListCmd.Flags().Int("page", 1, "page")
	versionsListCmd.Flags().Int("per-page", 20, "per page")
	versionsCreateCmd.Flags().String("space-id", "", "project/space id (required)")
	versionsCreateCmd.Flags().String("name", "", "version name (required)")
	versionsCreateCmd.Flags().String("owners", "", "comma-separated owner user ids (required)")
	versionsCreateCmd.Flags().String("start-date", "", "start date")
	versionsCreateCmd.Flags().String("publish-date", "", "publish date")
	versionsUpdateCmd.Flags().String("space-id", "", "project/space id (required)")
	versionsUpdateCmd.Flags().String("id", "", "version id (required)")
	versionsUpdateCmd.Flags().String("name", "", "version name (required)")
	versionsUpdateCmd.Flags().String("owners", "", "comma-separated owners")
	versionsUpdateCmd.Flags().String("start-date", "", "start date")
	versionsUpdateCmd.Flags().String("publish-date", "", "publish date")
	versionsDeleteCmd.Flags().String("space-id", "", "project/space id (required)")
	versionsDeleteCmd.Flags().String("id", "", "version id (required)")
	versionsCmd.AddCommand(versionsListCmd, versionsCreateCmd, versionsUpdateCmd, versionsDeleteCmd)
}
