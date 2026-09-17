package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
	"github.com/yunxiao-cli/yunxiao/internal/zhiyi"
)

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Projex sprints",
	Long: `Sprint typed commands (operations/projex/sprint.ts).

+shortcuts:
  yunxiao sprint +current   # aggregate recent Bug sprints (needs profile space_id)

Typed:
  yunxiao sprint list --space-id <projectId>
  yunxiao sprint get --space-id <id> --id <sprintId>
  yunxiao sprint create|update

Risk: +current/list/get=read; create/update=write`,
}

var sprintListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sprints in a project",
	Long:  "Risk: read\nHTTP: GET .../projects/{space}/sprints",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		spaceID, _ := cmd.Flags().GetString("space-id")
		status, _ := cmd.Flags().GetString("status")
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
		path, err := c.ProjexPath(cmd.Context(), "/projects/"+spaceID+"/sprints")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{}
		if status != "" {
			q["status"] = status
		}
		if page > 0 {
			q["page"] = strconv.Itoa(page)
		}
		if perPage > 0 {
			q["perPage"] = strconv.Itoa(perPage)
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var sprintGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a sprint",
	Long:  "Risk: read\nHTTP: GET .../projects/{space}/sprints/{id}",
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
		path, err := c.ProjexPath(cmd.Context(), "/projects/"+spaceID+"/sprints/"+id)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var sprintCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a sprint (write)",
	Long:  "Risk: write\nHTTP: POST .../projects/{space}/sprints\nBody: name, owners[], optional dates",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		spaceID, _ := cmd.Flags().GetString("space-id")
		name, _ := cmd.Flags().GetString("name")
		owners, _ := cmd.Flags().GetString("owners")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		description, _ := cmd.Flags().GetString("description")
		if err := requireFlags("space-id", spaceID, "name", name, "owners", owners); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/projects/"+spaceID+"/sprints")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"name": name, "owners": splitCSV(owners)}
		if startDate != "" {
			body["startDate"] = startDate
		}
		if endDate != "" {
			body["endDate"] = endDate
		}
		if description != "" {
			body["description"] = description
		}
		handleErr(runJSONMutating(cmd.Context(), c, "sprint create", risk.Write, "POST", path, nil, body, nil))
	},
}

var sprintUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a sprint (write)",
	Long:  "Risk: write\nHTTP: PUT .../projects/{space}/sprints/{id}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		spaceID, _ := cmd.Flags().GetString("space-id")
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		owners, _ := cmd.Flags().GetString("owners")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		description, _ := cmd.Flags().GetString("description")
		status, _ := cmd.Flags().GetString("status")
		if err := requireFlags("space-id", spaceID, "id", id); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/projects/"+spaceID+"/sprints/"+id)
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{}
		if name != "" {
			body["name"] = name
		}
		if owners != "" {
			body["owners"] = splitCSV(owners)
		}
		if startDate != "" {
			body["startDate"] = startDate
		}
		if endDate != "" {
			body["endDate"] = endDate
		}
		if description != "" {
			body["description"] = description
		}
		if status != "" {
			body["status"] = status
		}
		if len(body) == 0 {
			handleErr(fmt.Errorf("provide at least one field to update"))
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, "sprint update", risk.Write, "PUT", path, nil, body, nil))
	},
}

var sprintCurrentCmd = &cobra.Command{
	Use:   "+current",
	Short: "Shortcut: suggest current sprint from recent Bug work items",
	Long: `Risk: read

Needs active profile with space_id (e.g. --profile zhiyi).

POSTs workitems search (category=Bug, page=1, perPage=10) then aggregates sprint frequency.

  yunxiao sprint +current --profile zhiyi
  yunxiao sprint +current --profile zhiyi --dry-run

Success data: sampleSize, candidates[{id,name,count}], suggested.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		pf, err := requireProfile()
		if err != nil {
			handleErr(err)
			return
		}
		if pf.SpaceID == "" {
			handleErr(fmt.Errorf("profile %q missing space_id", pf.Name))
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems:search")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{
			"spaceId":  pf.SpaceID,
			"category": "Bug",
			"page":     1,
			"perPage":  10,
		}
		if globalDryRun {
			handleErr(output.DryRunResult(string(risk.Read), c.Preview("POST", path, nil, body)))
			return
		}
		var out any
		if err := c.Post(cmd.Context(), path, body, &out); err != nil {
			handleErr(err)
			return
		}
		result := zhiyi.AggregateBugSprints(zhiyi.ExtractSearchItems(out))
		handleErr(output.Success(result, map[string]any{"risk": risk.Read, "profile": pf.Name}))
	},
}

func init() {
	sprintListCmd.Flags().String("space-id", "", "project/space id (required)")
	sprintListCmd.Flags().String("status", "", "comma-separated status filter")
	sprintListCmd.Flags().Int("page", 1, "page")
	sprintListCmd.Flags().Int("per-page", 20, "per page")
	sprintGetCmd.Flags().String("space-id", "", "project/space id (required)")
	sprintGetCmd.Flags().String("id", "", "sprint id (required)")
	sprintCreateCmd.Flags().String("space-id", "", "project/space id (required)")
	sprintCreateCmd.Flags().String("name", "", "sprint name (required)")
	sprintCreateCmd.Flags().String("owners", "", "comma-separated owner user ids (required)")
	sprintCreateCmd.Flags().String("start-date", "", "start date")
	sprintCreateCmd.Flags().String("end-date", "", "end date")
	sprintCreateCmd.Flags().String("description", "", "description")
	sprintUpdateCmd.Flags().String("space-id", "", "project/space id (required)")
	sprintUpdateCmd.Flags().String("id", "", "sprint id (required)")
	sprintUpdateCmd.Flags().String("name", "", "name")
	sprintUpdateCmd.Flags().String("owners", "", "comma-separated owners")
	sprintUpdateCmd.Flags().String("start-date", "", "start date")
	sprintUpdateCmd.Flags().String("end-date", "", "end date")
	sprintUpdateCmd.Flags().String("description", "", "description")
	sprintUpdateCmd.Flags().String("status", "", "status")
	sprintCmd.AddCommand(sprintCurrentCmd, sprintListCmd, sprintGetCmd, sprintCreateCmd, sprintUpdateCmd)
}
