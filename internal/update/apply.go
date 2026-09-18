package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ExecutablePath resolves the running binary (symlink-evaluated when possible).
func ExecutablePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err2 := filepath.EvalSymlinks(path); err2 == nil {
		path = resolved
	}
	return path, nil
}

// DownloadFile downloads url to destPath (overwrites).
func DownloadFile(ctx context.Context, client HTTPDoer, url, destPath string) error {
	if client == nil {
		client = DefaultHTTPClient()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "yunxiao-cli-self-update")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}
	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, io.LimitReader(resp.Body, 256<<20)); err != nil {
		return err
	}
	return nil
}

// DownloadFirstOK tries candidates in order.
func DownloadFirstOK(ctx context.Context, client HTTPDoer, urls []string, destPath string) (used string, err error) {
	var errs []string
	for _, u := range urls {
		if u == "" {
			continue
		}
		if err := DownloadFile(ctx, client, u, destPath); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", u, err))
			_ = os.Remove(destPath)
			continue
		}
		fi, statErr := os.Stat(destPath)
		if statErr != nil || fi.Size() == 0 {
			errs = append(errs, fmt.Sprintf("%s: empty download", u))
			_ = os.Remove(destPath)
			continue
		}
		return u, nil
	}
	if len(errs) == 0 {
		return "", fmt.Errorf("no download URLs")
	}
	return "", fmt.Errorf("all downloads failed:\n  %s", strings.Join(errs, "\n  "))
}

// FileSHA256 returns hex sha256 of path.
func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyChecksum checks path against expected hex digest.
func VerifyChecksum(path, expectedHex string) error {
	expectedHex = strings.ToLower(strings.TrimSpace(expectedHex))
	if len(expectedHex) != 64 {
		return fmt.Errorf("invalid expected checksum")
	}
	actual, err := FileSHA256(path)
	if err != nil {
		return err
	}
	if actual != expectedHex {
		return fmt.Errorf("checksum mismatch: expected %s got %s", expectedHex, actual)
	}
	return nil
}

// ExtractBinary extracts BinaryName(platform) from archivePath into destDir; returns path.
func ExtractBinary(archivePath, destDir, platform string) (string, error) {
	want := BinaryName(platform)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return extractZipBinary(archivePath, destDir, want)
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return extractTarGzBinary(archivePath, destDir, want)
	default:
		return "", fmt.Errorf("unsupported archive format: %s", filepath.Base(archivePath))
	}
}

func extractTarGzBinary(archivePath, destDir, want string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		base := filepath.Base(hdr.Name)
		if base != want {
			continue
		}
		dest := filepath.Join(destDir, want)
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, io.LimitReader(tr, 256<<20)); err != nil {
			out.Close()
			return "", err
		}
		if err := out.Close(); err != nil {
			return "", err
		}
		return dest, nil
	}
	return "", fmt.Errorf("binary %s not found in archive", want)
}

func extractZipBinary(archivePath, destDir, want string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()
	for _, zf := range r.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(zf.Name) != want {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return "", err
		}
		dest := filepath.Join(destDir, want)
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			rc.Close()
			return "", err
		}
		_, copyErr := io.Copy(out, io.LimitReader(rc, 256<<20))
		rc.Close()
		closeErr := out.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		return dest, nil
	}
	return "", fmt.Errorf("binary %s not found in archive", want)
}

// ReplaceExecutable writes newBinary over targetPath safely:
// write beside as .new, rename target → .old, rename .new → target.
// On Windows the running image can be renamed; .old may remain until reboot.
func ReplaceExecutable(targetPath, newBinary string) error {
	targetPath = filepath.Clean(targetPath)
	newBinary = filepath.Clean(newBinary)
	dir := filepath.Dir(targetPath)
	base := filepath.Base(targetPath)
	staging := filepath.Join(dir, base+".new")
	backup := filepath.Join(dir, base+".old")

	// Copy to staging in the same directory (rename requires same volume).
	if err := copyFile(newBinary, staging); err != nil {
		return fmt.Errorf("stage new binary: %w", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(staging, 0o755); err != nil {
			_ = os.Remove(staging)
			return err
		}
	}

	_ = os.Remove(backup) // best-effort clear previous leftover
	if err := os.Rename(targetPath, backup); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("rename current → .old: %w (on Windows, close other yunxiao processes and retry)", err)
	}
	if err := os.Rename(staging, targetPath); err != nil {
		// try rollback
		_ = os.Rename(backup, targetPath)
		_ = os.Remove(staging)
		return fmt.Errorf("rename .new → current: %w", err)
	}
	_ = os.Remove(backup) // may fail on Windows while process still maps old image
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// VerifyBinaryVersion runs `path --version` and checks it contains wantVersion.
func VerifyBinaryVersion(path, wantVersion string) error {
	wantVersion = NormalizeVersion(wantVersion)
	cmd := exec.Command(path, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run %s --version: %w (%s)", path, err, truncate(string(out), 200))
	}
	if !strings.Contains(string(out), wantVersion) {
		return fmt.Errorf("updated binary reports %q, expected to contain %s", strings.TrimSpace(string(out)), wantVersion)
	}
	return nil
}

// FetchChecksums tries to download checksums.txt from the release assets.
func FetchChecksums(ctx context.Context, client HTTPDoer, rel *Release) (map[string]string, error) {
	asset, err := FindAsset(rel, "checksums.txt")
	if err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp("", "yunxiao-checksums-*.txt")
	if err != nil {
		return nil, err
	}
	path := tmp.Name()
	tmp.Close()
	defer os.Remove(path)
	urls := DownloadURLCandidates(asset.BrowserDownloadURL)
	if _, err := DownloadFirstOK(ctx, client, urls, path); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ChecksumMap(string(b)), nil
}
