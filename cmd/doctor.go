package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/client"
	"github.com/yunxiao-cli/yunxiao/internal/output"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "CLI health check: config, auth, and connectivity",
	Long:  "Risk: read",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		r, _, err := resolveEffectiveConfig()
		if err != nil {
			handleErr(err)
			return
		}
		tokenCheck := map[string]any{"name": "token", "ok": r.AccessToken != "", "source": r.TokenSource}
		if r.AccessToken == "" {
			tokenCheck["hint"] = "Create a Yunxiao personal access token: https://help.aliyun.com/zh/yunxiao/user-guide/personal-access-token"
		}
		checks := []map[string]any{
			{"name": "config_file", "ok": true, "path": r.ConfigPath},
			tokenCheck,
			{"name": "api_base_url", "ok": r.APIBaseURL != "", "value": r.APIBaseURL},
			{"name": "edition", "ok": r.Edition != "", "value": r.Edition},
		}
		connectivity := map[string]any{"name": "connectivity", "ok": false}
		if r.AccessToken != "" {
			c, err := client.New(r)
			if err == nil {
				var user map[string]any
				if err := c.Get(cmd.Context(), "/oapi/v1/platform/user", nil, &user); err != nil {
					connectivity["ok"] = false
					connectivity["error"] = err.Error()
				} else {
					connectivity["ok"] = true
					connectivity["user_id"] = user["id"]
					connectivity["user_name"] = user["name"]
					connectivity["last_organization"] = user["lastOrganization"]
				}
			} else {
				connectivity["error"] = err.Error()
			}
		} else {
			connectivity["error"] = "no token — create a PAT: https://help.aliyun.com/zh/yunxiao/user-guide/personal-access-token"
		}
		checks = append(checks, connectivity)
		allOK := true
		for _, ch := range checks {
			if ok, _ := ch["ok"].(bool); !ok {
				allOK = false
			}
		}
		meta := map[string]any{"healthy": allOK}
		if err := output.Success(checks, meta); err != nil {
			handleErr(err)
			return
		}
		if !allOK {
			handleErr(output.ExitError{Code: 1, Msg: "doctor found issues"})
		}
	},
}
