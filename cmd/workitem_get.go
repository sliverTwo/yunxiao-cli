package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
	"github.com/yunxiao-cli/yunxiao/internal/zhiyi"
)

var workitemGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a work item by ID or serial (e.g. ZYPT-5768)",
	Long:  "Risk: read\nHTTP: GET .../workitems/{id}\nAccepts --id or positional; Yunxiao accepts serial numbers like ZYPT-xxxx.",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		if _, err := applyActiveProfileOrg(); err != nil {
			handleErr(err)
			return
		}
		id, err := workitemIDFromFlagOrArg(cmd, args)
		if err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/workitems/"+id)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(runRead(cmd.Context(), c, "GET", path, nil, nil, map[string]any{"risk": risk.Read}, func(out any, meta map[string]any) (any, map[string]any) {
			zhiyi.EnrichWorkItemMeta(meta, asStringMap(out), profileSpaceID(), "")
			return out, meta
		}))
	},
}
