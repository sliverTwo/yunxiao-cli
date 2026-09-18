package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyArchiveChecksumFailsWhenArchiveIsMissingFromFetchedChecksums(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "yunxiao.tar.gz")
	if err := os.WriteFile(archivePath, []byte("archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	meta := map[string]any{}
	err := verifyArchiveChecksum(archivePath, filepath.Base(archivePath), map[string]string{}, nil, meta)
	if err == nil {
		t.Fatal("expected missing checksum error")
	}
	if !strings.Contains(err.Error(), "missing checksum") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := meta["checksum_ok"]; ok {
		t.Fatal("must not mark a missing checksum as soft-skipped")
	}
}

func TestVerifyArchiveChecksumSoftSkipsUnavailableChecksums(t *testing.T) {
	meta := map[string]any{}
	if err := verifyArchiveChecksum("unused", "yunxiao.tar.gz", nil, os.ErrNotExist, meta); err != nil {
		t.Fatal(err)
	}
	if got, ok := meta["checksum_ok"].(bool); !ok || got {
		t.Fatalf("checksum_ok = %#v, want false", meta["checksum_ok"])
	}
}
