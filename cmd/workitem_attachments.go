package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var workitemAttachmentsCmd = &cobra.Command{Use: "attachments", Short: "Work item file attachments"}

var workitemAttachmentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List attachments on a work item",
	Long:  "Risk: read\nHTTP: GET .../workitems/{id}/attachments\nSource: operations/projex/attachment.ts listWorkitemAttachmentsFunc",
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
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/attachments")
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, nil))
	},
}

var workitemAttachmentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Upload an attachment (multipart; write)",
	Long: `Risk: write
HTTP: POST multipart .../workitems/{id}/attachments  field "file"
Source: operations/projex/attachment.ts createWorkitemAttachmentFunc
--file must be a relative path under cwd (no absolute / no ..). Optional --file-name overrides basename.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		id, _ := cmd.Flags().GetString("id")
		file, _ := cmd.Flags().GetString("file")
		fileName, _ := cmd.Flags().GetString("file-name")
		operatorID, _ := cmd.Flags().GetString("operator-id")
		if err := requireFlags("id", id, "file", file); err != nil {
			handleErr(err)
			return
		}
		data, base, err := readRelativeFile(file)
		if err != nil {
			handleErr(err)
			return
		}
		if fileName == "" {
			fileName = base
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id+"/attachments")
		if err != nil {
			handleErr(err)
			return
		}
		fields := map[string]string{}
		if operatorID != "" {
			fields["operatorId"] = operatorID
		}
		preview := c.Preview("POST", path, nil, map[string]any{
			"multipart": true, "file_field": "file", "filename": fileName, "size": len(data), "fields": fields,
		})
		handleErr(runMutating("workitem attachments create", risk.Write, globalDryRun, globalYes, preview, func() error {
			var out any
			if err := c.PostMultipart(cmd.Context(), path, nil, "file", fileName, data, fields, &out); err != nil {
				return err
			}
			return output.Success(out, map[string]any{"risk": risk.Write})
		}))
	},
}
