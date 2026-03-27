package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"easy8-cli/internal/config"
	"easy8-cli/skills"
)

type commandInfo struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Flags       []flagInfo    `json:"flags,omitempty"`
	Subcommands []commandInfo `json:"subcommands,omitempty"`
}

type flagInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func runSkill(args []string, cfg config.Config) int {
	if len(args) == 0 {
		_, _ = os.Stdout.Write(skills.Content)
		if len(skills.Content) == 0 || skills.Content[len(skills.Content)-1] != '\n' {
			fmt.Fprintln(os.Stdout)
		}
		return 0
	}

	switch args[0] {
	case "install":
		return runSkillInstall(args[1:])
	case "help", "-h", "--help":
		printSkillUsage()
		return 0
	default:
		fmt.Fprintln(os.Stderr, "unknown skill command:", args[0])
		printSkillUsage()
		return 2
	}
}

func runSkillInstall(args []string) int {
	fs := flag.NewFlagSet("skill install", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	target := fs.String("target", "opencode", "Install target: opencode, claude, codex")
	pathFlag := fs.String("path", "", "Custom output path (file or directory)")
	global := fs.Bool("global", false, "Install into global agent directory")
	local := fs.Bool("local", false, "Install into local project directory")
	jsonOut := fs.Bool("json", false, "JSON envelope output")
	quietOut := fs.Bool("quiet", false, "Raw JSON data output")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := validateMachineFlags(*jsonOut, *quietOut); err != nil {
		return usageError(err)
	}
	if *global && *local {
		return usageError(fmt.Errorf("--global and --local cannot be used together"))
	}

	destination, scope, err := resolveSkillPath(strings.ToLower(strings.TrimSpace(*target)), strings.TrimSpace(*pathFlag), *global, *local)
	if err != nil {
		return usageError(err)
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if err := os.WriteFile(destination, skills.Content, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	data := map[string]any{
		"target": strings.ToLower(strings.TrimSpace(*target)),
		"scope":  scope,
		"path":   destination,
	}
	breadcrumbs := []outputBreadcrumb{
		{Action: "show", Cmd: "easy8 skill", Description: "Print installed skill"},
	}
	if *quietOut {
		return outputJSON(data)
	}
	if *jsonOut {
		return outputJSONEnvelope(data, "Skill installed", breadcrumbs, nil)
	}

	fmt.Fprintf(os.Stdout, "Skill installed to %s\n", destination)
	return 0
}

func resolveSkillPath(target, customPath string, global, local bool) (string, string, error) {
	if customPath != "" {
		if strings.HasSuffix(strings.ToLower(customPath), ".md") {
			return customPath, "custom", nil
		}
		return filepath.Join(customPath, "easy8-cli", "SKILL.md"), "custom", nil
	}

	scope := "global"
	if local {
		scope = "local"
	}
	if !global && !local {
		scope = "global"
	}

	if scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", "", err
		}
		switch target {
		case "opencode":
			return filepath.Join(home, ".config", "opencode", "skill", "easy8-cli", "SKILL.md"), scope, nil
		case "claude":
			return filepath.Join(home, ".claude", "skills", "easy8-cli", "SKILL.md"), scope, nil
		case "codex":
			return filepath.Join(home, ".codex", "skills", "easy8-cli", "SKILL.md"), scope, nil
		default:
			return "", "", fmt.Errorf("unsupported --target: %s", target)
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	switch target {
	case "opencode":
		return filepath.Join(wd, ".opencode", "skills", "easy8-cli", "SKILL.md"), scope, nil
	case "claude":
		return filepath.Join(wd, ".claude", "skills", "easy8-cli", "SKILL.md"), scope, nil
	case "codex":
		return filepath.Join(wd, ".codex", "skills", "easy8-cli", "SKILL.md"), scope, nil
	default:
		return "", "", fmt.Errorf("unsupported --target: %s", target)
	}
}

func printSkillUsage() {
	lines := []string{
		"easy8 skill",
		"",
		"Usage:",
		"  easy8 skill",
		"  easy8 skill install [flags]",
		"",
		"Examples:",
		"  easy8 skill",
		"  easy8 skill install --target opencode",
		"  easy8 skill install --target claude",
		"  easy8 skill install --target codex --local",
		"  easy8 skill install --path ./custom/skills",
	}
	for _, line := range lines {
		fmt.Fprintln(os.Stderr, line)
	}
}

func runCommands(args []string) int {
	fs := flag.NewFlagSet("commands", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	jsonOut := fs.Bool("json", false, "JSON envelope output")
	quietOut := fs.Bool("quiet", false, "Raw JSON data output")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := validateMachineFlags(*jsonOut, *quietOut); err != nil {
		return usageError(err)
	}

	catalog := buildCommandCatalog()
	if *quietOut {
		return outputJSON(catalog)
	}
	if *jsonOut {
		return outputJSONEnvelope(catalog, fmt.Sprintf("%d command groups", len(catalog)), nil, nil)
	}

	for _, cmd := range catalog {
		fmt.Fprintf(os.Stdout, "%s\t%s\n", cmd.Name, cmd.Description)
		for _, sub := range cmd.Subcommands {
			fmt.Fprintf(os.Stdout, "  %s\t%s\n", sub.Name, sub.Description)
		}
	}
	return 0
}

func buildCommandCatalog() []commandInfo {
	return []commandInfo{
		{
			Name:        "easy8 issue",
			Description: "Issue operations",
			Subcommands: []commandInfo{
				{Name: "easy8 issue create", Description: "Create issue"},
				{Name: "easy8 issue show", Description: "Show issue detail"},
				{Name: "easy8 issue list", Description: "List issues"},
				{Name: "easy8 issue search", Description: "Search issues"},
				{Name: "easy8 issue update", Description: "Update issue"},
			},
		},
		{
			Name:        "easy8 pbi",
			Description: "PBI operations",
			Subcommands: []commandInfo{
				{Name: "easy8 pbi list", Description: "List PBIs"},
				{Name: "easy8 pbi show", Description: "Show PBI detail"},
				{Name: "easy8 pbi update", Description: "Update PBI"},
			},
		},
		{
			Name:        "easy8 auth",
			Description: "Authentication helpers",
			Subcommands: []commandInfo{
				{Name: "easy8 auth status", Description: "Show auth status"},
				{Name: "easy8 auth login", Description: "Save API key"},
				{Name: "easy8 auth logout", Description: "Remove API key"},
			},
		},
		{
			Name:        "easy8 setup",
			Description: "Configure base URL and API key",
		},
		{
			Name:        "easy8 skill",
			Description: "Print/install skill file",
			Subcommands: []commandInfo{{Name: "easy8 skill install", Description: "Install skill for coding agents"}},
		},
		{
			Name:        "easy8 commands",
			Description: "List command catalog",
		},
		{
			Name:        "easy8 version",
			Description: "Show version",
		},
	}
}

func runAuth(args []string, cfg config.Config) int {
	if len(args) == 0 {
		printAuthUsage()
		return 2
	}

	switch args[0] {
	case "status":
		return runAuthStatus(args[1:], cfg)
	case "login":
		return runAuthLogin(args[1:], cfg)
	case "logout":
		return runAuthLogout(args[1:], cfg)
	case "help", "-h", "--help":
		printAuthUsage()
		return 0
	default:
		fmt.Fprintln(os.Stderr, "unknown auth command:", args[0])
		printAuthUsage()
		return 2
	}
}

func runAuthStatus(args []string, cfg config.Config) int {
	fs := flag.NewFlagSet("auth status", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	jsonOut := fs.Bool("json", false, "JSON envelope output")
	quietOut := fs.Bool("quiet", false, "Raw JSON data output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := validateMachineFlags(*jsonOut, *quietOut); err != nil {
		return usageError(err)
	}

	status := map[string]any{
		"authenticated": strings.TrimSpace(cfg.APIKey) != "",
		"base_url":      cfg.BaseURL,
	}
	if strings.TrimSpace(cfg.APIKey) != "" {
		status["api_key_masked"] = maskSecret(cfg.APIKey)
	}

	if *quietOut {
		return outputJSON(status)
	}
	if *jsonOut {
		summary := "Not authenticated"
		if strings.TrimSpace(cfg.APIKey) != "" {
			summary = "Authenticated"
		}
		return outputJSONEnvelope(status, summary, nil, nil)
	}

	fmt.Fprintf(os.Stdout, "Authenticated: %t\n", strings.TrimSpace(cfg.APIKey) != "")
	fmt.Fprintf(os.Stdout, "Base URL: %s\n", cfg.BaseURL)
	return 0
}

func runAuthLogin(args []string, cfg config.Config) int {
	fs := flag.NewFlagSet("auth login", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	apiKey := fs.String("api-key", "", "API key to save")
	local := fs.Bool("local", false, "Save into local .easy8.yaml")
	global := fs.Bool("global", false, "Save into global config")
	jsonOut := fs.Bool("json", false, "JSON envelope output")
	quietOut := fs.Bool("quiet", false, "Raw JSON data output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := validateMachineFlags(*jsonOut, *quietOut); err != nil {
		return usageError(err)
	}
	if *global && *local {
		return usageError(fmt.Errorf("--global and --local cannot be used together"))
	}

	key := strings.TrimSpace(*apiKey)
	if key == "" && fs.NArg() == 1 {
		key = strings.TrimSpace(fs.Arg(0))
	}
	if key == "" {
		return usageError(fmt.Errorf("API key is required (use --api-key or positional token)"))
	}

	cfg.APIKey = key
	var (
		path  string
		err   error
		scope = "global"
	)
	if *local {
		scope = "local"
		path, err = config.SaveLocal(cfg)
	} else {
		path, err = config.SaveGlobal(cfg)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		return 1
	}

	data := map[string]any{
		"authenticated": true,
		"scope":         scope,
		"path":          path,
	}
	breadcrumbs := []outputBreadcrumb{{Action: "status", Cmd: "easy8 auth status --json", Description: "Check auth status"}}
	if *quietOut {
		return outputJSON(data)
	}
	if *jsonOut {
		return outputJSONEnvelope(data, "API key saved", breadcrumbs, nil)
	}

	fmt.Fprintf(os.Stdout, "API key saved to %s\n", path)
	return 0
}

func runAuthLogout(args []string, cfg config.Config) int {
	fs := flag.NewFlagSet("auth logout", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	local := fs.Bool("local", false, "Clear API key in local .easy8.yaml")
	global := fs.Bool("global", false, "Clear API key in global config")
	jsonOut := fs.Bool("json", false, "JSON envelope output")
	quietOut := fs.Bool("quiet", false, "Raw JSON data output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := validateMachineFlags(*jsonOut, *quietOut); err != nil {
		return usageError(err)
	}
	if *global && *local {
		return usageError(fmt.Errorf("--global and --local cannot be used together"))
	}

	cfg.APIKey = ""
	var (
		path  string
		err   error
		scope = "global"
	)
	if *local {
		scope = "local"
		path, err = config.SaveLocal(cfg)
	} else {
		path, err = config.SaveGlobal(cfg)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		return 1
	}

	data := map[string]any{
		"authenticated": false,
		"scope":         scope,
		"path":          path,
	}
	breadcrumbs := []outputBreadcrumb{{Action: "login", Cmd: "easy8 auth login --api-key <key>", Description: "Save API key"}}
	if *quietOut {
		return outputJSON(data)
	}
	if *jsonOut {
		return outputJSONEnvelope(data, "API key removed", breadcrumbs, nil)
	}

	fmt.Fprintf(os.Stdout, "API key removed from %s\n", path)
	return 0
}

func printAuthUsage() {
	lines := []string{
		"easy8 auth",
		"",
		"Usage:",
		"  easy8 auth status [flags]",
		"  easy8 auth login [flags] [token]",
		"  easy8 auth logout [flags]",
	}
	for _, line := range lines {
		fmt.Fprintln(os.Stderr, line)
	}
}

func validateMachineFlags(jsonOut, quietOut bool) error {
	if jsonOut && quietOut {
		return fmt.Errorf("--json and --quiet cannot be used together")
	}
	return nil
}

func maskSecret(secret string) string {
	trimmed := strings.TrimSpace(secret)
	if len(trimmed) <= 6 {
		return strings.Repeat("*", len(trimmed))
	}
	return trimmed[:3] + strings.Repeat("*", len(trimmed)-6) + trimmed[len(trimmed)-3:]
}
