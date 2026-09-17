package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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

Risk: write (treat unknown endpoints carefully; use --dry-run first)`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		method := strings.ToUpper(args[0])
		path := args[1]
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		dataStr, _ := cmd.Flags().GetString("data")
		var body any
		if dataStr != "" {
			if err := json.Unmarshal([]byte(dataStr), &body); err != nil {
				handleErr(fmt.Errorf("--data must be JSON: %w", err))
				return
			}
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		if globalDryRun {
			handleErr(output.DryRunResult(string(risk.Write), c.Preview(method, path, nil, body)))
			return
		}
		// Mutating methods need --yes for high-risk gate when POST/PUT/PATCH/DELETE
		if method != "GET" && method != "HEAD" {
			if err := risk.CheckHighRisk("api "+method+" "+path, globalYes); err != nil {
				handleErr(err)
				return
			}
		}
		var out any
		if _, err := c.Do(cmd.Context(), method, path, nil, body, &out); err != nil {
			handleErr(err)
			return
		}
		handleErr(output.Success(out, map[string]any{"method": method, "path": path}))
	},
}

func init() {
	apiCmd.Flags().String("data", "", "JSON request body")
}
