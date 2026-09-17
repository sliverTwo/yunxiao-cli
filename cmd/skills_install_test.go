package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverInstallableSkills(t *testing.T) {
	root := t.TempDir()
	mustMkSkill := func(name string, withMD bool) {
		d := filepath.Join(root, name)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if withMD {
			if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	mustMkSkill("yunxiao-shared", true)
	mustMkSkill("yunxiao-codeup", true)
	mustMkSkill("yunxiao-empty", false)
	mustMkSkill("other-skill", true)
	if err := os.WriteFile(filepath.Join(root, "yunxiao-notadir"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	names, err := discoverInstallableSkills(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "yunxiao-codeup" || names[1] != "yunxiao-shared" {
		t.Fatalf("got %v", names)
	}

	names, err = discoverInstallableSkills(root, []string{"yunxiao-shared"})
	if err != nil || len(names) != 1 || names[0] != "yunxiao-shared" {
		t.Fatalf("filter: %v %v", names, err)
	}

	if _, err := discoverInstallableSkills(root, []string{"yunxiao-missing"}); err == nil {
		t.Fatal("expected missing skill error")
	}
}

func TestDefaultSkillsInstallDir(t *testing.T) {
	d := defaultSkillsInstallDir()
	if !filepath.IsAbs(d) && d != filepath.Join(".agents", "skills") {
		// When home is available, path should end with .agents/skills
	}
	if filepath.Base(d) != "skills" || filepath.Base(filepath.Dir(d)) != ".agents" {
		t.Fatalf("unexpected default dir %q", d)
	}
}

func TestInstallOneSkillCopyAndSkip(t *testing.T) {
	srcRoot := t.TempDir()
	dstRoot := t.TempDir()
	src := filepath.Join(srcRoot, "yunxiao-shared")
	if err := os.MkdirAll(filepath.Join(src, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "references", "a.md"), []byte("ref\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dstRoot, "yunxiao-shared")

	ins, sk, err := installOneSkill(src, dst, false, false, false)
	if err != nil || sk != nil || ins == nil || ins.Mode != "copy" {
		t.Fatalf("install: ins=%v sk=%v err=%v", ins, sk, err)
	}
	if _, err := os.Stat(filepath.Join(dst, "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, "references", "a.md")); err != nil {
		t.Fatal(err)
	}

	ins, sk, err = installOneSkill(src, dst, false, false, false)
	if err != nil || ins != nil || sk == nil {
		t.Fatalf("skip: ins=%v sk=%v err=%v", ins, sk, err)
	}

	ins, sk, err = installOneSkill(src, dst, false, true, true)
	if err != nil || sk != nil || ins == nil {
		t.Fatalf("dry-run force: ins=%v sk=%v err=%v", ins, sk, err)
	}
}
