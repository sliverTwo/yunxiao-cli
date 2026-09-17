package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/profile"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
	"github.com/yunxiao-cli/yunxiao/internal/zhiyi"
)

var workitemCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a work item",
	Long: `Risk: write
HTTP: POST .../workitems

When an active profile has workitem_defaults for --type-id, create fills
priority / trackers / 测试负责人 / 验收负责人 (and other defaults) unless
already set by flags / --custom-fields. Pass --no-defaults to skip.

If the API returns 未启用此字段【迭代】, omit --sprint for this workitem type
(Topic/Risk and some custom types do not enable 迭代).`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		spaceID, _ := cmd.Flags().GetString("space-id")
		typeID, _ := cmd.Flags().GetString("type-id")
		subject, _ := cmd.Flags().GetString("subject")
		assignedTo, _ := cmd.Flags().GetString("assigned-to")
		description, _ := cmd.Flags().GetString("description")
		formatType, _ := cmd.Flags().GetString("format-type")
		parentID, _ := cmd.Flags().GetString("parent-id")
		sprint, _ := cmd.Flags().GetString("sprint")
		labels, _ := cmd.Flags().GetString("labels")
		if err := requireFlags("space-id", spaceID, "type-id", typeID, "subject", subject, "assigned-to", assignedTo); err != nil {
			handleErr(err)
			return
		}
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
		body := map[string]any{
			"spaceId":        spaceID,
			"workitemTypeId": typeID,
			"subject":        subject,
			"assignedTo":     assignedTo,
		}
		if description != "" {
			body["description"] = description
		}
		if formatType != "" {
			body["formatType"] = formatType
		}
		if parentID != "" {
			body["parentId"] = parentID
		}
		if sprint != "" {
			body["sprint"] = sprint
		}
		if labels != "" {
			body["labels"] = splitCSV(labels)
		}
		participants, _ := cmd.Flags().GetString("participants")
		trackers, _ := cmd.Flags().GetString("trackers")
		verifier, _ := cmd.Flags().GetString("verifier")
		versions, _ := cmd.Flags().GetString("versions")
		if participants != "" {
			body["participants"] = splitCSV(participants)
		}
		if trackers != "" {
			body["trackers"] = splitCSV(trackers)
		}
		if verifier != "" {
			v, err := resolveSelfID(cmd.Context(), c, verifier)
			if err != nil {
				handleErr(err)
				return
			}
			body["verifier"] = v
		}
		if versions != "" {
			body["versions"] = splitCSV(versions)
		}
		cfJSON, _ := cmd.Flags().GetString("custom-fields")
		if cfJSON != "" {
			cf, err := parseJSONMap(cfJSON)
			if err != nil {
				handleErr(err)
				return
			}
			body["customFieldValues"] = cf
		}
		noDefaults, _ := cmd.Flags().GetBool("no-defaults")
		if !noDefaults {
			if pf, err := applyActiveProfileOrg(); err != nil {
				handleErr(err)
				return
			} else if pf != nil {
				profile.ApplyWorkitemDefaults(body, typeID, pf)
			}
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runJSONMutating(cmd.Context(), c, "workitem create", risk.Write, "POST", path, nil, body, func(out any, meta map[string]any) (any, map[string]any) {
			zhiyi.EnrichWorkItemMeta(meta, asStringMap(out), profileSpaceID(), spaceID)
			return out, meta
		}))
	},
}
