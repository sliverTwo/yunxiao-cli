package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/output"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Personal access token (PAT) management",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store Yunxiao personal access token in config file",
	Long: `Store a Yunxiao personal access token.

Create a PAT in the Yunxiao console (one-click):
  https://account-devops.aliyun.com/settings/personalAccessToken
Help: https://help.aliyun.com/zh/yunxiao/user-guide/personal-access-token

Token precedence: YUNXIAO_ACCESS_TOKEN env > active profile access_token > config.json.

Risk: write`,
	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		if token == "" {
			fmt.Fprint(os.Stderr, "Paste Yunxiao personal access token (input hidden not available; paste then Enter):\n> ")
			line, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				handleErr(err)
				return
			}
			token = strings.TrimSpace(line)
		}
		if token == "" {
			fmt.Fprintf(os.Stderr, "Create a Yunxiao PAT (console):\n  %s\nHelp: %s\n%s\n", yunxiaoPATConsoleURL, yunxiaoPATHelpURL, patPermissionsGuide())
			handleErr(output.Fail(output.ErrorBody{
				Type:    "cli",
				Message: "empty token",
				Hint:    patHintShort(),
			}, 1))
			return
		}
		f, _, err := config.LoadFile()
		if err != nil {
			handleErr(err)
			return
		}
		f.AccessToken = token
		if base, _ := cmd.Flags().GetString("api-base-url"); base != "" {
			f.APIBaseURL = strings.TrimRight(base, "/")
		}
		if org, _ := cmd.Flags().GetString("organization-id"); org != "" {
			f.OrganizationID = org
		}
		p, err := config.SaveFile(f)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(output.Success(map[string]any{
			"config_path":  p,
			"token_masked": config.MaskToken(token),
			"hint":         "token stored with mode 0600; prefer YUNXIAO_ACCESS_TOKEN in CI",
		}, nil))
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show whether a token is configured (never prints the raw token)",
	Long:  "Risk: read",
	Run: func(cmd *cobra.Command, args []string) {
		flagOrg(globalOrg)
		r, pf, err := resolveEffectiveConfig()
		if err != nil {
			handleErr(err)
			return
		}
		out := map[string]any{
			"has_token":       r.AccessToken != "",
			"token_source":    r.TokenSource,
			"token_masked":    config.MaskToken(r.AccessToken),
			"api_base_url":    r.APIBaseURL,
			"organization_id": r.OrganizationID,
			"edition":         r.Edition,
			"config_path":     r.ConfigPath,
		}
		if pf != nil {
			out["profile"] = pf.Name
		}
		handleErr(output.Success(out, nil))
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove access token from config file",
	Long:  "Risk: write",
	Run: func(cmd *cobra.Command, args []string) {
		f, _, err := config.LoadFile()
		if err != nil {
			handleErr(err)
			return
		}
		f.AccessToken = ""
		p, err := config.SaveFile(f)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(output.Success(map[string]any{"config_path": p, "cleared": true}, nil))
	},
}

func init() {
	authLoginCmd.Flags().String("token", "", "PAT value (otherwise prompted)")
	authLoginCmd.Flags().String("api-base-url", "", "override API base URL")
	authLoginCmd.Flags().String("organization-id", "", "default organization ID")
	authCmd.AddCommand(authLoginCmd, authStatusCmd, authLogoutCmd)
}
