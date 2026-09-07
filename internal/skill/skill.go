// Package skill は同梱の SKILL.md（AI エージェント向けの使い方）。
package skill

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed SKILL.md
var Content string

// ClaudePath は Claude Code のスキルの置き場所（~/.claude/skills/pitboard/SKILL.md）
func ClaudePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".claude", "skills", "pitboard", "SKILL.md")
}

func InstallClaude() (string, error) {
	path := ClaudePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, []byte(Content), 0o644)
}
