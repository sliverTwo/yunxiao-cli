package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/profile"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Tenant profiles (org/space/bug workflow constants)",
	Long: `Load tenant-specific Projex constants from ~/.config/yunxiao/profiles/<name>.json.

  yunxiao profile list
  yunxiao profile show [name]
  yunxiao profile path [name]
  yunxiao profile doctor [name] [--all-workflows]
  yunxiao profile install-example zhiyi|play [--force]

Select with --profile <name> or YUNXIAO_PROFILE=<name>.
Ship examples: profiles/zhiyi.example.json (full Zhiyi fields), profiles/play.example.json (sandbox-minimal).`,
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed profiles",
	Long:  "Risk: read",
	Run: func(cmd *cobra.Command, args []string) {
		names, dir, err := profile.ListNames()
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(output.Success(map[string]any{
			"profiles":     names,
			"profiles_dir": dir,
			"active":       activeProfileName(),
		}, map[string]any{"risk": risk.Read}))
	},
}

var profileShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show a profile (default: active --profile / YUNXIAO_PROFILE)",
	Long:  "Risk: read",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) > 0 {
			name = args[0]
		} else {
			name = activeProfileName()
		}
		if name == "" {
			handleErr(profile.HintMissing())
			return
		}
		p, err := profile.Load(name)
		if err != nil {
			handleErr(err)
			return
		}
		path, _ := profile.Path(name)
		// Never print raw PAT from profile show.
		if p.AccessToken != "" {
			p.AccessToken = config.MaskToken(p.AccessToken)
		}
		handleErr(output.Success(p, map[string]any{"risk": risk.Read, "path": path}))
	},
}

var profilePathCmd = &cobra.Command{
	Use:   "path [name]",
	Short: "Print profiles dir or a profile file path",
	Long:  "Risk: read",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			d, err := profile.Dir()
			if err != nil {
				handleErr(err)
				return
			}
			handleErr(output.Success(map[string]any{"path": d}, map[string]any{"risk": risk.Read}))
			return
		}
		p, err := profile.Path(args[0])
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(output.Success(map[string]any{"path": p}, map[string]any{"risk": risk.Read}))
	},
}

var profileInstallExampleCmd = &cobra.Command{
	Use:   "install-example <name>",
	Short: "Copy profiles/<name>.example.json into ~/.config/yunxiao/profiles/",
	Long:  "Risk: write\nExample: yunxiao profile install-example zhiyi",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")
		src, err := findProfileExample(name)
		if err != nil {
			handleErr(err)
			return
		}
		preview := map[string]any{"from": src, "name": name, "force": force}
		if globalDryRun {
			dst, _ := profile.Path(name)
			preview["to"] = dst
			handleErr(output.DryRunResult(string(risk.Write), preview))
			return
		}
		dst, err := profile.InstallExample(name, src, force)
		if err != nil {
			handleErr(err)
			return
		}
		handleErr(output.Success(map[string]any{
			"installed": dst,
			"hint":      fmt.Sprintf("example uses placeholders only — fill org/space/type IDs via profile set / explore-workflow; then: export YUNXIAO_PROFILE=%s  # or --profile %s", name, name),
		}, map[string]any{"risk": risk.Write}))
	},
}

func findProfileExample(name string) (string, error) {
	filename := name + ".example.json"
	candidates := []string{}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates,
			filepath.Join(filepath.Dir(exe), "profiles", filename),
			filepath.Join(filepath.Dir(exe), "..", "profiles", filename),
		)
	}
	_, file, _, _ := runtime.Caller(0)
	candidates = append(candidates, filepath.Join(filepath.Dir(file), "..", "profiles", filename))
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "profiles", filename))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs, nil
		}
	}
	return "", fmt.Errorf("example not found: profiles/%s (searched under repo/binary)", filename)
}

func init() {
	profileInstallExampleCmd.Flags().Bool("force", false, "overwrite existing profile")
	profileCmd.AddCommand(profileListCmd, profileShowCmd, profilePathCmd, profileInstallExampleCmd)
}
