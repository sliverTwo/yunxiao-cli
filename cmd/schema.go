package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/schema"
)

var schemaCmd = &cobra.Command{
	Use:   "schema [id]",
	Short: "View API method parameters, types, and risk",
	Long: `Inspect a typed method before calling it.

Examples:
  yunxiao schema
  yunxiao schema codeup.mrs.create
  yunxiao schema --domain project

Risk: read`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain, _ := cmd.Flags().GetString("domain")
		if len(args) == 0 {
			handleErr(output.Success(schema.List(domain), map[string]any{"count": len(schema.List(domain))}))
			return
		}
		id := strings.TrimSpace(args[0])
		m := schema.Find(id)
		if m == nil {
			handleErr(fmt.Errorf("unknown schema id %q; try: yunxiao schema", id))
			return
		}
		handleErr(output.Success(m, nil))
	},
}

func init() {
	schemaCmd.Flags().String("domain", "", "filter by domain")
}
