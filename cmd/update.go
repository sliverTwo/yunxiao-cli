package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
	"github.com/yunxiao-cli/yunxiao/internal/update"
	"github.com/yunxiao-cli/yunxiao/internal/version"
)

var updateCheckOnly bool

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"self-update"},
	Short:   "Update the yunxiao binary from GitHub Releases",
	Long: `Risk: write (replaces the local yunxiao executable)

Compares the running version to the latest GitHub Release, downloads the
matching platform archive (windows/linux/darwin × amd64/arm64), verifies
SHA-256 when checksums.txt is present, and replaces this binary safely
(write beside → rename; Windows-friendly).

  yunxiao update --check          # report only; exit 2 if update available
  yunxiao update --dry-run        # same as --check
  yunxiao update                  # TTY: confirm; non-TTY: requires --yes
  yunxiao update --yes            # apply without prompt (scripts/CI)

No network on other commands by default. Optional: yunxiao doctor --check-update.
Disable optional hints: YUNXIAO_UPDATE_CHECK=0.

Override release source (same as npm installer):
  YUNXIAO_CLI_GITHUB_REPO=owner/repo
  YUNXIAO_CLI_DOWNLOAD_BASE=https://example.com/path

npm installs can also refresh via: npm install -g sanzhi-yunxiao-cli@latest`,
	Run: func(cmd *cobra.Command, args []string) {
		handleErr(runUpdate(cmd))
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "only check for a newer release (no download); exit 2 if available")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command) error {
	checkOnly := updateCheckOnly || globalDryRun

	platform, arch, err := update.CurrentPlatformArch()
	if err != nil {
		return err
	}
	exe, err := update.ExecutablePath()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	client := update.DefaultHTTPClient()
	repo := update.GithubRepo()
	rel, err := update.FetchLatestRelease(cmd.Context(), client, repo)
	if err != nil {
		return err
	}
	latest := update.NormalizeVersion(rel.TagName)
	current := update.NormalizeVersion(version.Version)
	newer := update.NewerAvailable(current, latest)

	meta := map[string]any{
		"current":          current,
		"latest":           latest,
		"tag":              rel.TagName,
		"repo":             repo,
		"platform":         platform,
		"arch":             arch,
		"executable":       exe,
		"release_url":      rel.HTMLURL,
		"update_available": newer,
	}

	if checkOnly {
		data := map[string]any{
			"status":  update.FormatPair(current, latest),
			"message": update.FormatPair(current, latest),
		}
		if err := output.Success(data, meta); err != nil {
			return err
		}
		if newer {
			return output.ExitError{Code: update.ExitUpdateAvailable, Msg: "update available"}
		}
		return nil
	}

	if !newer {
		return output.Success(map[string]any{
			"status":  "already_latest",
			"message": update.FormatPair(current, latest),
		}, meta)
	}

	archiveName := update.ArchiveName(latest, platform, arch)
	asset, err := update.FindAsset(rel, archiveName)
	if err != nil {
		return err
	}
	meta["archive"] = archiveName

	if err := confirmSelfUpdate(latest, exe, globalYes); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "yunxiao-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archiveName)
	urls := update.ArchiveDownloadCandidates(archiveName, asset.BrowserDownloadURL)
	usedURL, err := update.DownloadFirstOK(cmd.Context(), client, urls, archivePath)
	if err != nil {
		return err
	}
	meta["download_url"] = usedURL

	if sums, sumErr := update.FetchChecksums(cmd.Context(), client, rel); sumErr == nil {
		if expected, ok := sums[archiveName]; ok {
			if err := update.VerifyChecksum(archivePath, expected); err != nil {
				return err
			}
			meta["checksum_ok"] = true
		} else {
			meta["checksum_ok"] = false
			meta["checksum_note"] = "archive not listed in checksums.txt"
		}
	} else {
		meta["checksum_ok"] = false
		meta["checksum_note"] = "checksums.txt not available; skipped verify"
	}

	extractDir := filepath.Join(tmpDir, "extract")
	binPath, err := update.ExtractBinary(archivePath, extractDir, platform)
	if err != nil {
		return err
	}
	if err := update.ReplaceExecutable(exe, binPath); err != nil {
		return err
	}
	if err := update.VerifyBinaryVersion(exe, latest); err != nil {
		meta["verify_warning"] = err.Error()
		meta["hint"] = "binary replaced; restart the process if --version still shows the old build (common on Windows)"
	} else {
		meta["verified_version"] = latest
	}

	return output.Success(map[string]any{
		"status":  "updated",
		"message": fmt.Sprintf("updated %s → %s", current, latest),
		"path":    exe,
	}, meta)
}

func confirmSelfUpdate(latest, exe string, yes bool) error {
	if yes {
		return nil
	}
	if stdinIsInteractive() {
		fmt.Fprintf(os.Stderr, "Update yunxiao to %s?\n  replace: %s\nConfirm [y/N]: ", latest, exe)
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return err
		}
		ans := strings.ToLower(strings.TrimSpace(line))
		if ans == "y" || ans == "yes" {
			return nil
		}
		return output.Fail(output.ErrorBody{
			Type:    "cli",
			Message: "update cancelled",
		}, 1)
	}
	return risk.CheckConfirmed("yunxiao update (replace local binary)", risk.Write, false)
}
