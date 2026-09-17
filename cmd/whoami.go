package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/client"
	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/output"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current effective identity and token status (JSON)",
	Long:  "Risk: read",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		r, pf, err := resolveEffectiveConfig()
		if err != nil {
			handleErr(err)
			return
		}
		result := map[string]any{
			"token_source":    r.TokenSource,
			"token_masked":    config.MaskToken(r.AccessToken),
			"api_base_url":    r.APIBaseURL,
			"organization_id": r.OrganizationID,
			"edition":         r.Edition,
			"config_path":     r.ConfigPath,
		}
		if pf != nil {
			result["profile"] = pf.Name
		}
		if r.AccessToken != "" {
			c, err := client.New(r)
			if err == nil {
				var user map[string]any
				if err := c.Get(cmd.Context(), "/oapi/v1/platform/user", nil, &user); err == nil {
					result["user"] = user
				} else {
					result["user_error"] = err.Error()
				}
			} else {
				result["client_error"] = err.Error()
			}
		}
		handleErr(output.Success(result, nil))
	},
}
