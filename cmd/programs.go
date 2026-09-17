package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var programsCmd = &cobra.Command{
	Use:   "programs",
	Short: "Projex programs (project sets)",
	Long: `Search programs / project sets (operations/projex/project.ts searchProgramsFunc).

  yunxiao programs search [--name …] [--status …] [--page 1]

Risk: read`,
}

var programsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search programs",
	Long:  "Risk: read\nHTTP: POST .../programs:search\nSource: searchProgramsFunc",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		name, _ := cmd.Flags().GetString("name")
		status, _ := cmd.Flags().GetString("status")
		gmtStart, _ := cmd.Flags().GetString("gmt-create-start")
		gmtEnd, _ := cmd.Flags().GetString("gmt-create-end")
		creator, _ := cmd.Flags().GetString("creator")
		users, _ := cmd.Flags().GetString("users")
		orderBy, _ := cmd.Flags().GetString("order-by")
		sort, _ := cmd.Flags().GetString("sort")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.ProjexPath(cmd.Context(), "/programs:search")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{
			"page":    page,
			"perPage": perPage,
			"orderBy": orderBy,
			"sort":    sort,
		}
		if name != "" {
			body["name"] = name
		}
		if status != "" {
			body["status"] = status
		}
		if gmtStart != "" {
			body["gmtCreateStart"] = gmtStart
		}
		if gmtEnd != "" {
			body["gmtCreateEnd"] = gmtEnd
		}
		if creator != "" {
			body["creator"] = creator
		}
		if users != "" {
			body["users"] = users
		}
		handleErr(runRead(cmd.Context(), c, "POST", path, nil, body, map[string]any{
			"risk": risk.Read,
			"page": strconv.Itoa(page),
		}, nil))
	},
}

func init() {
	programsSearchCmd.Flags().String("name", "", "fuzzy name")
	programsSearchCmd.Flags().String("status", "", "status ids (comma-separated)")
	programsSearchCmd.Flags().String("gmt-create-start", "", "yyyy-MM-dd HH:mm:ss")
	programsSearchCmd.Flags().String("gmt-create-end", "", "yyyy-MM-dd HH:mm:ss")
	programsSearchCmd.Flags().String("creator", "", "creator ids (comma-separated)")
	programsSearchCmd.Flags().String("users", "", "user ids for extraConditions")
	programsSearchCmd.Flags().String("order-by", "gmtCreate", "gmtCreate|name")
	programsSearchCmd.Flags().String("sort", "desc", "asc|desc")
	programsSearchCmd.Flags().Int("page", 1, "page")
	programsSearchCmd.Flags().Int("per-page", 20, "per page (0-200)")
	programsCmd.AddCommand(programsSearchCmd)
	rootCmd.AddCommand(programsCmd)
}
