package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var workitemEffortsCmd = &cobra.Command{
	Use:   "efforts",
	Short: "Actual effort records",
	Long: `Actual effort records (operations/projex/effort.ts).

  yunxiao workitem efforts mine --start-date 2026-01-01 --end-date 2026-01-31
  yunxiao workitem efforts list --id <workitemId>
  yunxiao workitem efforts create --id <id> --actual-time 2 --gmt-start … --gmt-end … --dry-run
  yunxiao workitem efforts update --workitem-id <wid> --id <rid> … --dry-run

Risk: list/mine=read; create/update=write`,
}

var workitemEstimatedCmd = &cobra.Command{
	Use:   "estimated-efforts",
	Short: "Estimated efforts",
	Long: `Estimated efforts (operations/projex/effort.ts).

  yunxiao workitem estimated-efforts list --id <workitemId>
  yunxiao workitem estimated-efforts create --id <id> --owner <uid> --spent-time 4 --dry-run
  yunxiao workitem estimated-efforts update --workitem-id <wid> --id <eid> … --dry-run

Risk: list=read; create/update=write`,
}

var workitemEffortsMineCmd = &cobra.Command{
	Use:   "mine",
	Short: "List current user effort records",
	Long:  "Risk: read\nHTTP: GET .../effortRecords?startDate=&endDate=",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		start, _ := cmd.Flags().GetString("start-date")
		end, _ := cmd.Flags().GetString("end-date")
		if err := requireFlags("start-date", start, "end-date", end); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/effortRecords")
		if err != nil {
			handleErr(err)
			return
		}
		q := map[string]string{"startDate": start, "endDate": end}
		handleErr(runRead(cmd.Context(), c, "GET", path, q, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var workitemEffortsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List effort records for a work item",
	Long:  "Risk: read\nHTTP: GET .../workitems/{id}/effortRecords\n\nDefault order: newest first by update/create time. Use --sort asc for oldest first.",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		id, _ := cmd.Flags().GetString("id")
		if err := requireFlags("id", id); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/effortRecords")
		if err != nil {
			handleErr(err)
			return
		}
		sortFlag, _ := cmd.Flags().GetString("sort")
		after, err := afterSortByTime(sortFlag, nil)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, after))
	},
}

var workitemEffortsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an actual effort record",
	Long:  "Risk: write\nHTTP: POST .../workitems/{id}/effortRecords",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		id, _ := cmd.Flags().GetString("id")
		actual, _ := cmd.Flags().GetFloat64("actual-time")
		gmtStart, _ := cmd.Flags().GetString("gmt-start")
		gmtEnd, _ := cmd.Flags().GetString("gmt-end")
		desc, _ := cmd.Flags().GetString("description")
		op, _ := cmd.Flags().GetString("operator-id")
		workType, _ := cmd.Flags().GetString("work-type")
		if err := requireFlags("id", id, "gmt-start", gmtStart, "gmt-end", gmtEnd); err != nil {
			handleErr(err)
			return
		}
		if actual <= 0 {
			handleErr(requireFlags("actual-time", ""))
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/effortRecords")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"actualTime": actual, "gmtStart": gmtStart, "gmtEnd": gmtEnd}
		if desc != "" {
			body["description"] = desc
		}
		if op != "" {
			body["operatorId"] = op
		}
		if workType != "" {
			body["workType"] = workType
		}
		handleErr(runJSONMutating(cmd.Context(), c, "workitem efforts create", risk.Write, "POST", path, nil, body, nil))
	},
}

var workitemEffortsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an actual effort record",
	Long:  "Risk: write\nHTTP: PUT .../workitems/{workitemId}/effortRecords/{id}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		wid, _ := cmd.Flags().GetString("workitem-id")
		id, _ := cmd.Flags().GetString("id")
		actual, _ := cmd.Flags().GetFloat64("actual-time")
		gmtStart, _ := cmd.Flags().GetString("gmt-start")
		gmtEnd, _ := cmd.Flags().GetString("gmt-end")
		desc, _ := cmd.Flags().GetString("description")
		op, _ := cmd.Flags().GetString("operator-id")
		workType, _ := cmd.Flags().GetString("work-type")
		if err := requireFlags("workitem-id", wid, "id", id, "gmt-start", gmtStart, "gmt-end", gmtEnd); err != nil {
			handleErr(err)
			return
		}
		if actual <= 0 {
			handleErr(requireFlags("actual-time", ""))
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+wid+"/effortRecords/"+id)
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"actualTime": actual, "gmtStart": gmtStart, "gmtEnd": gmtEnd}
		if desc != "" {
			body["description"] = desc
		}
		if op != "" {
			body["operatorId"] = op
		}
		if workType != "" {
			body["workType"] = workType
		}
		handleErr(runJSONMutating(cmd.Context(), c, "workitem efforts update", risk.Write, "PUT", path, nil, body, nil))
	},
}

var workitemEstimatedListCmd = &cobra.Command{
	Use:   "list",
	Short: "List estimated efforts for a work item",
	Long:  "Risk: read\nHTTP: GET .../workitems/{id}/estimatedEfforts\n\nDefault order: newest first by update/create time. Use --sort asc for oldest first.",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		id, _ := cmd.Flags().GetString("id")
		if err := requireFlags("id", id); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/estimatedEfforts")
		if err != nil {
			handleErr(err)
			return
		}
		sortFlag, _ := cmd.Flags().GetString("sort")
		after, err := afterSortByTime(sortFlag, nil)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, after))
	},
}

var workitemEstimatedCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an estimated effort",
	Long:  "Risk: write\nHTTP: POST .../workitems/{id}/estimatedEfforts",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		id, _ := cmd.Flags().GetString("id")
		owner, _ := cmd.Flags().GetString("owner")
		spent, _ := cmd.Flags().GetFloat64("spent-time")
		desc, _ := cmd.Flags().GetString("description")
		op, _ := cmd.Flags().GetString("operator-id")
		workType, _ := cmd.Flags().GetString("work-type")
		if err := requireFlags("id", id, "owner", owner); err != nil {
			handleErr(err)
			return
		}
		if spent <= 0 {
			handleErr(requireFlags("spent-time", ""))
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		ownerID, err := resolveSelfID(cmd.Context(), c, owner)
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/estimatedEfforts")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"owner": ownerID, "spentTime": spent}
		if desc != "" {
			body["description"] = desc
		}
		if op != "" {
			body["operatorId"] = op
		}
		if workType != "" {
			body["workType"] = workType
		}
		handleErr(runJSONMutating(cmd.Context(), c, "workitem estimated-efforts create", risk.Write, "POST", path, nil, body, nil))
	},
}

var workitemEstimatedUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an estimated effort",
	Long:  "Risk: write\nHTTP: PUT .../workitems/{workitemId}/estimatedEfforts/{id}",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		wid, _ := cmd.Flags().GetString("workitem-id")
		id, _ := cmd.Flags().GetString("id")
		owner, _ := cmd.Flags().GetString("owner")
		spent, _ := cmd.Flags().GetFloat64("spent-time")
		desc, _ := cmd.Flags().GetString("description")
		op, _ := cmd.Flags().GetString("operator-id")
		workType, _ := cmd.Flags().GetString("work-type")
		if err := requireFlags("workitem-id", wid, "id", id, "owner", owner); err != nil {
			handleErr(err)
			return
		}
		if spent <= 0 {
			handleErr(requireFlags("spent-time", ""))
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		ownerID, err := resolveSelfID(cmd.Context(), c, owner)
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+wid+"/estimatedEfforts/"+id)
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"owner": ownerID, "spentTime": spent}
		if desc != "" {
			body["description"] = desc
		}
		if op != "" {
			body["operatorId"] = op
		}
		if workType != "" {
			body["workType"] = workType
		}
		handleErr(runJSONMutating(cmd.Context(), c, "workitem estimated-efforts update", risk.Write, "PUT", path, nil, body, nil))
	},
}

func init() {
	workitemEffortsMineCmd.Flags().String("start-date", "", "yyyy-MM-dd (required)")
	workitemEffortsMineCmd.Flags().String("end-date", "", "yyyy-MM-dd (required)")
	workitemEffortsListCmd.Flags().String("id", "", "work item id (required)")
	addSortFlag(workitemEffortsListCmd)
	workitemEffortsCreateCmd.Flags().String("id", "", "work item id (required)")
	workitemEffortsCreateCmd.Flags().Float64("actual-time", 0, "actual hours (required, >0)")
	workitemEffortsCreateCmd.Flags().String("gmt-start", "", "start date (required)")
	workitemEffortsCreateCmd.Flags().String("gmt-end", "", "end date (required)")
	workitemEffortsCreateCmd.Flags().String("description", "", "description")
	workitemEffortsCreateCmd.Flags().String("operator-id", "", "operator user id")
	workitemEffortsCreateCmd.Flags().String("work-type", "", "work type")
	workitemEffortsUpdateCmd.Flags().String("workitem-id", "", "work item id (required)")
	workitemEffortsUpdateCmd.Flags().String("id", "", "effort record id (required)")
	workitemEffortsUpdateCmd.Flags().Float64("actual-time", 0, "actual hours (required, >0)")
	workitemEffortsUpdateCmd.Flags().String("gmt-start", "", "start date (required)")
	workitemEffortsUpdateCmd.Flags().String("gmt-end", "", "end date (required)")
	workitemEffortsUpdateCmd.Flags().String("description", "", "description")
	workitemEffortsUpdateCmd.Flags().String("operator-id", "", "operator user id")
	workitemEffortsUpdateCmd.Flags().String("work-type", "", "work type")
	workitemEstimatedListCmd.Flags().String("id", "", "work item id (required)")
	addSortFlag(workitemEstimatedListCmd)
	workitemEstimatedCreateCmd.Flags().String("id", "", "work item id (required)")
	workitemEstimatedCreateCmd.Flags().String("owner", "", "owner user id or self (required)")
	workitemEstimatedCreateCmd.Flags().Float64("spent-time", 0, "estimated hours (required, >0)")
	workitemEstimatedCreateCmd.Flags().String("description", "", "description")
	workitemEstimatedCreateCmd.Flags().String("operator-id", "", "operator user id")
	workitemEstimatedCreateCmd.Flags().String("work-type", "", "work type")
	workitemEstimatedUpdateCmd.Flags().String("workitem-id", "", "work item id (required)")
	workitemEstimatedUpdateCmd.Flags().String("id", "", "estimated effort id (required)")
	workitemEstimatedUpdateCmd.Flags().String("owner", "", "owner user id or self (required)")
	workitemEstimatedUpdateCmd.Flags().Float64("spent-time", 0, "estimated hours (required, >0)")
	workitemEstimatedUpdateCmd.Flags().String("description", "", "description")
	workitemEstimatedUpdateCmd.Flags().String("operator-id", "", "operator user id")
	workitemEstimatedUpdateCmd.Flags().String("work-type", "", "work type")
	workitemEffortsCmd.AddCommand(workitemEffortsMineCmd, workitemEffortsListCmd, workitemEffortsCreateCmd, workitemEffortsUpdateCmd)
	workitemEstimatedCmd.AddCommand(workitemEstimatedListCmd, workitemEstimatedCreateCmd, workitemEstimatedUpdateCmd)
	workitemCmd.AddCommand(workitemEffortsCmd, workitemEstimatedCmd)
}
