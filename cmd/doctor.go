package cmd

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/client"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/profile"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "CLI health check: config, auth, and connectivity",
	Long: `Risk: read

Prints resolved executable path (os.Executable / argv0; Windows-friendly),
active profile summary (name, organization_id, space_id), config/token checks,
and a connectivity probe.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		r, pf, err := resolveEffectiveConfig()
		if err != nil {
			handleErr(err)
			return
		}
		tokenCheck := map[string]any{"name": "token", "ok": r.AccessToken != "", "source": r.TokenSource, "token_kind": string(r.TokenKind)}
		if r.AccessToken == "" {
			tokenCheck["hint"] = "Prefer: yunxiao auth login --browser — or PAT: " + patHintShort()
			tokenCheck["console"] = yunxiaoPATConsoleURL
			tokenCheck["browser"] = "yunxiao auth login --browser"
		} else if r.TokenKind != "" {
			tokenCheck["auth_header"] = r.AuthHeader
		}
		checks := []map[string]any{
			executableCheck(),
			activeProfileCheck(pf),
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
			connectivity["error"] = "no token — run: yunxiao auth login --browser (or --token / YUNXIAO_ACCESS_TOKEN)"
		}
		checks = append(checks, connectivity)
		allOK := true
		for _, ch := range checks {
			if ok, _ := ch["ok"].(bool); !ok {
				allOK = false
			}
		}
		meta := map[string]any{"healthy": allOK, "goos": runtime.GOOS, "goarch": runtime.GOARCH}
		if err := output.Success(checks, meta); err != nil {
			handleErr(err)
			return
		}
		if !allOK {
			handleErr(output.ExitError{Code: 1, Msg: "doctor found issues"})
		}
	},
}

func executableCheck() map[string]any {
	argv0 := ""
	if len(os.Args) > 0 {
		argv0 = os.Args[0]
	}
	path, err := os.Executable()
	ch := map[string]any{
		"name":  "executable",
		"ok":    true,
		"argv0": argv0,
	}
	if err != nil {
		ch["path"] = argv0
		ch["executable_error"] = err.Error()
		return ch
	}
	if resolved, err2 := filepath.EvalSymlinks(path); err2 == nil {
		path = resolved
	}
	ch["path"] = path
	return ch
}

func activeProfileCheck(pf *profile.Profile) map[string]any {
	name := activeProfileName()
	ch := map[string]any{
		"name":          "active_profile",
		"ok":            true,
		"profile_name":  name,
		"profile_set":   name != "",
	}
	if pf == nil {
		if name == "" {
			ch["hint"] = "no active profile (set YUNXIAO_PROFILE or --profile)"
		}
		return ch
	}
	ch["organization_id"] = pf.OrganizationID
	ch["space_id"] = pf.SpaceID
	return ch
}
