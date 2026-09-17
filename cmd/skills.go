package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yunxiao-cli/yunxiao/internal/output"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "List, read, or install embedded agent skill content",
}

func skillsRoot() string {
	// Prefer skills/ next to binary's module when developing; else relative to this source.
	candidates := []string{}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "skills"))
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "..", "skills"))
	}
	_, file, _, _ := runtime.Caller(0)
	candidates = append(candidates, filepath.Join(filepath.Dir(file), "..", "skills"))
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "skills"))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return "skills"
}

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available agent skills",
	Long:  "Risk: read",
	Run: func(cmd *cobra.Command, args []string) {
		root := skillsRoot()
		entries, err := os.ReadDir(root)
		if err != nil {
			handleErr(err)
			return
		}
		var names []string
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "yunxiao-") {
				names = append(names, e.Name())
			}
		}
		handleErr(output.Success(names, map[string]any{"skills_dir": root}))
	},
}

var skillsReadCmd = &cobra.Command{
	Use:   "read <name>",
	Short: "Read a skill SKILL.md",
	Long:  "Risk: read",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		root := skillsRoot()
		name := args[0]
		p := filepath.Join(root, name, "SKILL.md")
		b, err := os.ReadFile(p)
		if err != nil {
			handleErr(fmt.Errorf("skill %q not found at %s: %w", name, p, err))
			return
		}
		handleErr(output.Success(map[string]any{"name": name, "path": p, "content": string(b)}, nil))
	},
}

var skillsPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print skills directory path (for agent install)",
	Long:  "Risk: read",
	Run: func(cmd *cobra.Command, args []string) {
		handleErr(output.Success(map[string]any{"path": skillsRoot()}, nil))
	},
}

func init() {
	skillsCmd.AddCommand(skillsListCmd, skillsReadCmd, skillsPathCmd, skillsInstallCmd)
}
