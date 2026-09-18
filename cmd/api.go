package cmd

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/client"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var apiCmd = &cobra.Command{
	Use:   "api <METHOD> <path>",
	Short: "Raw HTTP escape hatch — call any Yunxiao OpenAPI path",
	Long: `Raw escape hatch when no typed command exists.

Examples:
  yunxiao api GET /oapi/v1/platform/user
  yunxiao api POST /oapi/v1/projex/organizations/{org}/workitems:search --data '{"category":"Req"}'
  yunxiao api POST /oapi/v1/... --data-file body.json
  yunxiao api POST /oapi/v1/... --data @body.json

Risk: GET/HEAD and known read POSTs (paths containing ":search", e.g. workitems:search)
are read (no --yes). Other POST/PUT/PATCH/DELETE are high-risk-write and need --yes;
use --dry-run first.

For :search / list-style responses, meta includes pagination from x-* headers
(pagination_headers, has_more, total, …) when the API returns them.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		method := strings.ToUpper(args[0])
		path := args[1]
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")
		body, err := loadJSONBodyFromFlags(dataStr, dataFile)
		if err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		if globalDryRun {
			rl := risk.Write
			if risk.IsReadOnlyHTTP(method, path) {
				rl = risk.Read
			}
			handleErr(output.DryRunResult(string(rl), c.Preview(method, path, nil, body)))
			return
		}
		// Mutating methods need --yes. Genuine read POSTs (e.g. workitems:search)
		// are Risk: read and must not require confirmation.
		if !risk.IsReadOnlyHTTP(method, path) {
			if err := risk.CheckHighRisk("api "+method+" "+path, globalYes); err != nil {
				handleErr(err)
				return
			}
		}
		var out any
		hdr, err := c.Do(cmd.Context(), method, path, nil, body, &out)
		if err != nil {
			handleErr(err)
			return
		}
		meta := map[string]any{"method": method, "path": path}
		if risk.IsReadOnlyHTTP(method, path) || strings.Contains(path, ":search") {
			meta = client.MetaWithPagination(meta, hdr)
			if bodyMap, ok := body.(map[string]any); ok {
				if cond, ok := bodyMap["conditions"]; ok {
					meta["request"] = map[string]any{"conditions": cond, "body_keys": mapKeys(bodyMap)}
				}
			}
		}
		handleErr(output.Success(out, meta))
	},
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func init() {
	apiCmd.Flags().String("data", "", "JSON request body (or @file.json)")
	apiCmd.Flags().String("data-file", "", "read JSON body from file (alternative to --data)")
}
