package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/client"
	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/profile"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

// stdinIsInteractive reports whether stdin is a TTY (tests may override).
var stdinIsInteractive = func() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// onboardStdin is the reader used for numbered project selection (tests may redirect).
var onboardStdin io.Reader = os.Stdin

// onboardStderr is where interactive prompts are printed (tests may redirect).
var onboardStderr io.Writer = os.Stderr

type onboardProject struct {
	ID   string
	Name string
}

var onboardCmd = &cobra.Command{
	Use:   "+onboard",
	Short: "Interactive local profile setup: pick a project, write ~/.config/yunxiao/profiles/<name>.json",
	Long: `Risk: write (local config only — never writes into the git working tree)

Checks auth, lists Projex projects, then writes a generic local profile
(space_id + name only) under the local config profiles directory.

TTY: numbered project select.
Non-TTY: require --space-id (optional --name / --profile).

  yunxiao +onboard
  yunxiao +onboard --space-id <id> --profile myproj
  yunxiao +onboard --space-id <id> --dry-run

Do NOT use this to embed 智衣/沙箱 tenant templates into a repo.
Real tenant specifics stay on the user's machine (untracked local profiles).

Auth: prefer yunxiao auth login --browser (OAuth = full account API capability).
PAT fallback: ` + yunxiaoPATConsoleURL + `
Help: ` + yunxiaoPATHelpURL,
	Run: func(cmd *cobra.Command, args []string) {
		handleErr(runOnboard(cmd))
	},
}

func errNeedAuth() error {
	msg := "not authenticated: run yunxiao auth login --browser (or --token <PAT> / YUNXIAO_ACCESS_TOKEN)"
	fmt.Fprintln(onboardStderr, "Recommended: yunxiao auth login --browser")
	fmt.Fprintln(onboardStderr, "WARNING: OAuth consent = full account API capability (no module scopes).")
	fmt.Fprintf(onboardStderr, "PAT fallback (console):\n  %s\n", yunxiaoPATConsoleURL)
	fmt.Fprintf(onboardStderr, "Help: %s\n", yunxiaoPATHelpURL)
	fmt.Fprintln(onboardStderr, patPermissionsGuide())
	return output.Fail(output.ErrorBody{
		Type:    "cli",
		Message: msg,
		Hint:    "yunxiao auth login --browser — or PAT: " + patHintShort(),
	}, 1)
}

func runOnboard(cmd *cobra.Command) error {
	flagOrg(globalOrg)
	spaceID, _ := cmd.Flags().GetString("space-id")
	nameFlag, _ := cmd.Flags().GetString("name")
	profileFlag, _ := cmd.Flags().GetString("profile")
	force, _ := cmd.Flags().GetBool("force")

	r, _, err := resolveEffectiveConfig()
	if err != nil {
		return err
	}
	if r.AccessToken == "" {
		return errNeedAuth()
	}

	c, _, err := mustClient()
	if err != nil {
		return err
	}

	var projects []onboardProject
	var selectedName string

	spaceID = strings.TrimSpace(spaceID)
	if spaceID == "" {
		projects, err = fetchOnboardProjects(cmd.Context(), c)
		if err != nil {
			return err
		}
		if len(projects) == 0 {
			return fmt.Errorf("no projects returned; pass --space-id explicitly")
		}
		if !stdinIsInteractive() {
			return fmt.Errorf("non-interactive session: pass --space-id <projectId> (optional --name / --profile)")
		}
		picked, err := promptOnboardProject(projects)
		if err != nil {
			return err
		}
		spaceID = picked.ID
		selectedName = picked.Name
	} else if nameFlag == "" && profileFlag == "" {
		// Best-effort: resolve display name from list when only space-id is given.
		if list, listErr := fetchOnboardProjects(cmd.Context(), c); listErr == nil {
			for _, p := range list {
				if p.ID == spaceID {
					selectedName = p.Name
					break
				}
			}
			projects = list
		}
	}

	profileName := resolveOnboardProfileName(profileFlag, nameFlag, selectedName, spaceID)
	displayName := strings.TrimSpace(nameFlag)
	if displayName == "" {
		displayName = selectedName
	}
	if displayName == "" {
		displayName = profileName
	}

	dst, err := profile.Path(profileName)
	if err != nil {
		return err
	}
	profilesDir, err := profile.Dir()
	if err != nil {
		return err
	}

	exists := false
	if st, err := os.Stat(dst); err == nil && !st.IsDir() {
		exists = true
	}

	plan := map[string]any{
		"action":          "write_local_profile",
		"profile_name":    profileName,
		"space_id":        spaceID,
		"display_name":    displayName,
		"path":            dst,
		"profiles_dir":    profilesDir,
		"exists":          exists,
		"force":           force,
		"organization_id": r.OrganizationID,
		"token_source":    r.TokenSource,
		"token_masked":    config.MaskToken(r.AccessToken),
		"hint":            "writes only under local config dir; never into the git working tree",
		"next": []string{
			fmt.Sprintf("export YUNXIAO_PROFILE=%s", profileName),
			fmt.Sprintf("yunxiao profile show %s", profileName),
			fmt.Sprintf("yunxiao profile doctor %s", profileName),
		},
	}
	if len(projects) > 0 {
		plan["projects_listed"] = len(projects)
	}

	if globalDryRun {
		return output.DryRunResult(string(risk.Write), plan)
	}

	pf, err := buildOnboardProfile(profileName, spaceID, r.OrganizationID, force)
	if err != nil {
		return err
	}
	saved, err := pf.Save()
	if err != nil {
		return err
	}

	return output.Success(map[string]any{
		"profile_name":    profileName,
		"space_id":        spaceID,
		"name":            pf.Name,
		"path":            saved,
		"profiles_dir":    profilesDir,
		"updated":         exists,
		"organization_id": strings.TrimSpace(r.OrganizationID),
		"export":          fmt.Sprintf("export YUNXIAO_PROFILE=%s", profileName),
		"suggest": []string{
			fmt.Sprintf("yunxiao profile show %s", profileName),
			fmt.Sprintf("yunxiao profile doctor %s", profileName),
			"yunxiao doctor",
		},
		"note": "generic local profile (space_id + name). Tenant-specific fields (智衣/沙箱) stay on this machine — do not commit them to the public repo.",
	}, map[string]any{"risk": risk.Write})
}

func fetchOnboardProjects(ctx context.Context, c *client.Client) ([]onboardProject, error) {
	path, err := c.ProjexPath(ctx, "/projects:search")
	if err != nil {
		return nil, err
	}
	var body any
	if err := c.Post(ctx, path, map[string]any{"page": 1, "perPage": 100}, &body); err != nil {
		return nil, err
	}
	return parseOnboardProjects(body), nil
}

func parseOnboardProjects(body any) []onboardProject {
	items := client.ExtractListItems(body)
	var out []onboardProject
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		id := firstNonEmptyString(m, "id", "identifier", "spaceId", "space_id")
		name := firstNonEmptyString(m, "name", "displayName", "display_name", "customCode", "identifier")
		if id == "" {
			continue
		}
		if name == "" {
			name = id
		}
		out = append(out, onboardProject{ID: id, Name: name})
	}
	return out
}

func firstNonEmptyString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		switch v := m[k].(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		case float64:
			return strconv.FormatInt(int64(v), 10)
		}
	}
	return ""
}

func promptOnboardProject(projects []onboardProject) (onboardProject, error) {
	fmt.Fprintln(onboardStderr, "Select a Projex project for a LOCAL profile (written under ~/.config/yunxiao/profiles/):")
	for i, p := range projects {
		fmt.Fprintf(onboardStderr, "  %d. %s  (%s)\n", i+1, p.Name, p.ID)
	}
	fmt.Fprint(onboardStderr, "Enter number: ")
	line, err := bufio.NewReader(onboardStdin).ReadString('\n')
	if err != nil && len(strings.TrimSpace(line)) == 0 {
		return onboardProject{}, fmt.Errorf("read selection: %w", err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || n < 1 || n > len(projects) {
		return onboardProject{}, fmt.Errorf("invalid selection %q (want 1..%d)", strings.TrimSpace(line), len(projects))
	}
	return projects[n-1], nil
}

func resolveOnboardProfileName(profileFlag, nameFlag, projectName, spaceID string) string {
	if s := strings.TrimSpace(profileFlag); s != "" {
		return sanitizeProfileName(s)
	}
	if s := strings.TrimSpace(nameFlag); s != "" {
		return sanitizeProfileName(s)
	}
	if s := sanitizeProfileName(projectName); s != "" {
		return s
	}
	if s := sanitizeProfileName(spaceID); s != "" {
		if len(s) > 24 {
			s = s[:24]
		}
		return s
	}
	return "default"
}

func sanitizeProfileName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		case unicode.IsSpace(r):
			b.WriteByte('-')
		default:
			// drop non-ASCII (incl. CJK) for portable file names;
			// pass --profile explicitly when the project name is not ASCII.
		}
	}
	out := strings.Trim(b.String(), "-._")
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return out
}

// buildOnboardProfile creates or updates a generic local profile (space_id + name).
// With force=true, replaces the file with a minimal template (drops tenant fields).
// Without force, merges into an existing profile when present.
func buildOnboardProfile(profileName, spaceID, orgID string, force bool) (*profile.Profile, error) {
	profileName = strings.TrimSpace(profileName)
	if profileName == "" {
		return nil, fmt.Errorf("empty profile name")
	}
	spaceID = strings.TrimSpace(spaceID)
	if spaceID == "" {
		return nil, fmt.Errorf("empty space_id")
	}

	var pf *profile.Profile
	if !force {
		if existing, err := profile.Load(profileName); err == nil {
			pf = existing
		}
	}
	if pf == nil {
		pf = &profile.Profile{}
	}
	// Profile.Name is the logical profile name (file basename).
	pf.Name = profileName
	pf.SpaceID = spaceID
	if org := strings.TrimSpace(orgID); org != "" {
		pf.OrganizationID = org
	}
	// Never copy access_token from env into the profile file during onboard.
	return pf, nil
}

func init() {
	onboardCmd.Flags().String("space-id", "", "Projex project/space id (required when non-TTY)")
	onboardCmd.Flags().String("name", "", "human display hint used to derive profile filename when --profile omitted")
	onboardCmd.Flags().String("profile", "", "local profile name (file ~/.config/yunxiao/profiles/<name>.json)")
	onboardCmd.Flags().Bool("force", false, "replace existing profile with a minimal template (drop extra fields)")
	rootCmd.AddCommand(onboardCmd)
}
