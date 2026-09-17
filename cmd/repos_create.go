package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var codeupReposCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Codeup repository (high-risk-write)",
	Long: `Create repository (operations/codeup/repositories.ts createRepositoryFunc).

  yunxiao codeup repos create --name my-repo --path my-repo [--visibility private] --dry-run

Query always sends createParentPath=true (MCP default).
Risk: high-risk-write`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		name, _ := cmd.Flags().GetString("name")
		pathName, _ := cmd.Flags().GetString("path")
		desc, _ := cmd.Flags().GetString("description")
		ns, _ := cmd.Flags().GetInt("namespace-id")
		vis, _ := cmd.Flags().GetString("visibility")
		avatar, _ := cmd.Flags().GetString("avatar-url")
		readme, _ := cmd.Flags().GetString("readme-type")
		if err := requireFlags("name", name, "path", pathName); err != nil {
			handleErr(err)
			return
		}
		c, _, err := mustClient()
		if err != nil {
			handleErr(err)
			return
		}
		path, err := c.CodeupPath(cmd.Context(), "/repositories")
		if err != nil {
			handleErr(err)
			return
		}
		body := map[string]any{"name": name, "path": pathName}
		if desc != "" {
			body["description"] = desc
		}
		if ns > 0 {
			body["namespaceId"] = ns
		}
		if vis != "" {
			body["visibility"] = vis
		}
		if avatar != "" {
			body["avatarUrl"] = avatar
		}
		if readme != "" {
			body["readMeType"] = readme
		}
		q := map[string]string{"createParentPath": "true"}
		handleErr(runJSONMutating(cmd.Context(), c, "codeup repos create", risk.HighRiskWrite, "POST", path, q, body, nil))
	},
}

func init() {
	codeupReposCreateCmd.Flags().String("name", "", "repository name (required)")
	codeupReposCreateCmd.Flags().String("path", "", "repository path (required)")
	codeupReposCreateCmd.Flags().String("description", "", "description")
	codeupReposCreateCmd.Flags().Int("namespace-id", 0, "parent namespace id")
	codeupReposCreateCmd.Flags().String("visibility", "", "private|internal")
	codeupReposCreateCmd.Flags().String("avatar-url", "", "avatar URL")
	codeupReposCreateCmd.Flags().String("readme-type", "", "EMPTY|USER_GUIDE")
	codeupReposCmd.AddCommand(codeupReposCreateCmd)
}
