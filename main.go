package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type branchInfo struct {
	Name string `json:"name"`
	Env  string `json:"env"`
}

type scanResult struct {
	Hash     string        `json:"hash"`
	Message  string        `json:"message"`
	Author   string        `json:"author"`
	Date     string        `json:"date"`
	Branches []branchInfo  `json:"branches"`
	Pipeline []pipelineEnv `json:"pipeline"`
}

type pipelineEnv struct {
	Name     string   `json:"name"`
	Branches []string `json:"branches"`
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFB86C")).
			MarginTop(1)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8BE9FD")).
			Width(12)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	hashStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#50FA7B"))

	branchNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Bold(true)

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	successIcon = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	errorIcon = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	otherStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#6272A4")).
			Padding(0, 1)
)

func makeEnvStyle(color string) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color(color)).
		Padding(0, 1)
}

func getEnvStyleByName(name string, cfg Config) lipgloss.Style {
	for _, env := range cfg.Environments {
		if env.Name == name {
			return makeEnvStyle(env.Color)
		}
	}
	return otherStyle
}

type model struct {
	spinner  spinner.Model
	message  string
	done     bool
	result   scanResult
	err      error
	jsonOut  bool
	shortOut bool
	config   Config
}

func newSpinnerModel(msg string, jsonOut, shortOut bool, cfg Config) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	return model{
		spinner:  s,
		message:  msg,
		jsonOut:  jsonOut,
		shortOut: shortOut,
		config:   cfg,
	}
}

type scanCompleteMsg struct {
	result scanResult
	err    error
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.scanCommit())
}

func (m model) scanCommit() tea.Cmd {
	return func() tea.Msg {
		commitHash := m.message

		if err := verifyCommit(commitHash); err != nil {
			return scanCompleteMsg{err: err}
		}

		commitInfo, err := getCommitInfo(commitHash)
		if err != nil {
			return scanCompleteMsg{err: fmt.Errorf("failed to get commit info: %w", err)}
		}

		branches, err := findBranchesWithCommit(commitHash)
		if err != nil {
			return scanCompleteMsg{err: fmt.Errorf("failed to find branches: %w", err)}
		}

		parts := strings.Split(commitInfo, "|")
		result := scanResult{
			Hash:    parts[0],
			Message: safeIndex(parts, 1),
			Author:  safeIndex(parts, 2),
			Date:    safeIndex(parts, 3),
		}

		for _, b := range branches {
			env := matchEnvironment(b, m.config)
			result.Branches = append(result.Branches, branchInfo{
				Name: b,
				Env:  env,
			})
		}

		envMap := categorizeBranches(branches, m.config)
		for _, env := range m.config.Environments {
			if bs, ok := envMap[env.Name]; ok {
				result.Pipeline = append(result.Pipeline, pipelineEnv{
					Name:     env.Name,
					Branches: bs,
				})
			}
		}
		if bs, ok := envMap["OTHER"]; ok {
			result.Pipeline = append(result.Pipeline, pipelineEnv{
				Name:     "OTHER",
				Branches: bs,
			})
		}

		return scanCompleteMsg{result: result}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case scanCompleteMsg:
		m.done = true
		m.result = msg.result
		m.err = msg.err
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.done {
		if m.err != nil {
			return renderError(m.err)
		}
		if m.jsonOut {
			return renderJSON(m.result)
		}
		if m.shortOut {
			return renderShort(m.result)
		}
		return renderFull(m.result, m.config)
	}

	return fmt.Sprintf("\n  %s %s\n", m.spinner.View(), dimStyle.Render("Scanning remote branches..."))
}

func main() {
	args := os.Args[1:]
	jsonOut := false
	shortOut := false
	initMode := false
	webMode := false
	upgradeMode := false
	checkMode := false
	versionsMode := false
	webAddr := ":9876"
	commitHash := ""
	installVersion := ""

	for i, arg := range args {
		switch arg {
		case "--help", "-h":
			printHelp()
			return
		case "--version", "-v":
			printVersion()
			return
		case "--json", "-j":
			jsonOut = true
		case "--short", "-s":
			shortOut = true
		case "--init":
			initMode = true
		case "--web":
			webMode = true
		case "--upgrade":
			upgradeMode = true
		case "--check":
			checkMode = true
		case "--versions":
			versionsMode = true
		case "--install":
			if i+1 < len(args) {
				installVersion = args[i+1]
			}
		default:
			if !strings.HasPrefix(arg, "-") && installVersion == "" {
				commitHash = arg
			}
		}
	}

	if initMode {
		if err := initConfig(); err != nil {
			fmt.Fprintln(os.Stderr, errorIcon.Render(" Error: ")+err.Error())
			os.Exit(1)
		}
		fmt.Println(successIcon.Render(" Created .code-radar.yaml"))
		fmt.Println(dimStyle.Render(" Edit it to match your branch naming conventions"))
		return
	}

	if upgradeMode {
		runUpgrade()
		return
	}

	if checkMode {
		runCheck()
		return
	}

	if versionsMode {
		runVersions()
		return
	}

	if installVersion != "" {
		runInstall(installVersion)
		return
	}

	if webMode {
		startWebServer(webAddr)
		return
	}

	if commitHash == "" {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("  📡 Code Radar"))
		fmt.Println(dimStyle.Render("  Track where your git commits have been deployed"))
		fmt.Println()
		fmt.Print("  Enter commit hash: ")
		fmt.Scanln(&commitHash)
	}

	commitHash = strings.TrimSpace(commitHash)
	if commitHash == "" {
		fmt.Fprintln(os.Stderr, errorIcon.Render(" Error: ")+"commit hash is required")
		os.Exit(1)
	}

	cfg := loadConfig()

	if !isTTY() {
		renderDirect(commitHash, jsonOut, shortOut, cfg)
		return
	}

	m := newSpinnerModel(commitHash, jsonOut, shortOut, cfg)
	p := tea.NewProgram(m, tea.WithOutput(os.Stdout))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, errorIcon.Render(" Error: ")+err.Error())
		os.Exit(1)
	}
}

func renderError(err error) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(errorIcon.Render(" Error: ") + err.Error())
	b.WriteString("\n\n")
	return b.String()
}

func renderJSON(r scanResult) string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

func renderShort(r scanResult) string {
	var b strings.Builder
	hash := r.Hash
	if len(hash) > 12 {
		hash = hash[:12]
	}
	b.WriteString(fmt.Sprintf("%s ", hashStyle.Render(hash)))
	if r.Message != "" {
		b.WriteString(valueStyle.Render(truncateString(r.Message, 40)) + " ")
	}
	if len(r.Branches) > 0 {
		envs := make([]string, 0)
		seen := make(map[string]bool)
		for _, br := range r.Branches {
			if !seen[br.Env] {
				seen[br.Env] = true
				envs = append(envs, br.Env)
			}
		}
		b.WriteString(dimStyle.Render("→ " + strings.Join(envs, " → ")))
	}
	b.WriteString("\n")
	return b.String()
}

func renderFull(r scanResult, cfg Config) string {
	var b strings.Builder
	hash := r.Hash
	if len(hash) > 12 {
		hash = hash[:12]
	}

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  📡  CODE RADAR  "))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render("  ┌─ COMMIT DETAILS"))
	b.WriteString("\n")
	b.WriteString("  │\n")
	b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Hash:"), hashStyle.Render(hash)))
	if r.Message != "" {
		b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Message:"), valueStyle.Render(truncateString(r.Message, 55))))
	}
	if r.Author != "" {
		b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Author:"), valueStyle.Render(r.Author)))
	}
	if r.Date != "" {
		b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Date:"), dateStyle.Render(truncateString(r.Date, 40))))
	}
	b.WriteString("  │\n\n")

	if len(r.Branches) == 0 {
		b.WriteString(headerStyle.Render("  ┌─ BRANCHES"))
		b.WriteString("\n")
		b.WriteString("  │\n")
		b.WriteString(fmt.Sprintf("  %s %s\n", errorIcon.Render(" ✗ "), errorStyle().Render("No remote branches contain this commit")))
		b.WriteString("  │\n")
	} else {
		b.WriteString(headerStyle.Render(fmt.Sprintf("  ┌─ BRANCHES (%d found)", len(r.Branches))))
		b.WriteString("\n")
		b.WriteString("  │\n")

		for i, br := range r.Branches {
			connector := "├"
			if i == len(r.Branches)-1 {
				connector = "└"
			}

			envStyle := getEnvStyleByName(br.Env, cfg)
			envTag := envStyle.Render(fmt.Sprintf(" %s ", br.Env))

			b.WriteString(fmt.Sprintf("  %s %s %s %s\n",
				dimStyle.Render(connector+"─"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Render("→"),
				envTag,
				branchNameStyle.Render(br.Name),
			))
		}
		b.WriteString("  │\n")
	}

	b.WriteString("\n")
	return b.String()
}

func errorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555")).
		Bold(true)
}

func safeIndex(parts []string, i int) string {
	if i < len(parts) {
		return parts[i]
	}
	return ""
}

func verifyCommit(hash string) error {
	cmd := exec.Command("git", "cat-file", "-t", hash)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("commit '%s' not found", hash)
	}
	return nil
}

func getCommitInfo(hash string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%H|%s|%an|%ai", hash)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func findBranchesWithCommit(hash string) ([]string, error) {
	fetchCmd := exec.Command("git", "fetch", "--all", "--quiet")
	fetchCmd.Run()

	cmd := exec.Command("git", "branch", "-r", "--contains", hash, "--format=%(refname:short)")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	rawBranches := parseBranches(string(out))
	seen := make(map[string]bool)
	var branches []string

	for _, b := range rawBranches {
		if isRemoteOnly(b) {
			continue
		}
		name := stripRemotePrefix(b)
		if !seen[name] {
			seen[name] = true
			branches = append(branches, name)
		}
	}

	sort.Strings(branches)
	return branches, nil
}

func parseBranches(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches
}

func stripRemotePrefix(branch string) string {
	parts := strings.SplitN(branch, "/", 2)
	if len(parts) == 2 {
		remote := parts[0]
		if remote == "origin" || remote == "upstream" || remote == "remote" {
			return parts[1]
		}
	}
	return branch
}

func isRemoteOnly(branch string) bool {
	return !strings.Contains(branch, "/")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func printHelp() {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("  📡 Code Radar"))
	fmt.Println(dimStyle.Render("  Track where your git commits have been deployed"))
	fmt.Println()

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
	commandStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	flagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))

	fmt.Println(helpStyle.Render("  Usage:"))
	fmt.Println()
	fmt.Printf("    %s %s\n", commandStyle.Render("code-radar"), flagStyle.Render("[flags] <commit-hash>"))
	fmt.Println()

	fmt.Println(helpStyle.Render("  Scan Flags:"))
	fmt.Println()
	fmt.Printf("    %s, %-10s %s\n", flagStyle.Render("--help"), flagStyle.Render("-h"), helpStyle.Render("Show this help message"))
	fmt.Printf("    %s, %-10s %s\n", flagStyle.Render("--version"), flagStyle.Render("-v"), helpStyle.Render("Show version information"))
	fmt.Printf("    %s, %-10s %s\n", flagStyle.Render("--json"), flagStyle.Render("-j"), helpStyle.Render("Output as JSON"))
	fmt.Printf("    %s, %-10s %s\n", flagStyle.Render("--short"), flagStyle.Render("-s"), helpStyle.Render("Compact output"))
	fmt.Printf("    %s        %s\n", flagStyle.Render("--init"), helpStyle.Render("Generate .code-radar.yaml config"))
	fmt.Printf("    %s        %s\n", flagStyle.Render("--web"), helpStyle.Render("Start web UI on localhost:9876"))
	fmt.Println()

	fmt.Println(helpStyle.Render("  Update Flags:"))
	fmt.Println()
	fmt.Printf("    %s          %s\n", flagStyle.Render("--upgrade"), helpStyle.Render("Upgrade to latest version"))
	fmt.Printf("    %s          %s\n", flagStyle.Render("--check"), helpStyle.Render("Check for latest version"))
	fmt.Printf("    %s        %s\n", flagStyle.Render("--versions"), helpStyle.Render("List available versions"))
	fmt.Printf("    %s      %s\n", flagStyle.Render("--install"), helpStyle.Render("Install specific version (e.g. v1.0.0)"))
	fmt.Println()

	fmt.Println(helpStyle.Render("  Config:"))
	fmt.Println()
	fmt.Println(helpStyle.Render("    Create .code-radar.yaml in your project root:"))
	fmt.Println()
	fmt.Println(dimStyle.Render("      environments:"))
	fmt.Println(dimStyle.Render("        - name: DEV"))
	fmt.Println(dimStyle.Render("          pattern: [develop, dev]"))
	fmt.Println(dimStyle.Render("          color: \"#6272A4\""))
	fmt.Println()
	fmt.Println(helpStyle.Render("    Run code-radar --init to generate a default config"))
	fmt.Println()
}

func printVersion() {
	fmt.Println()
	fmt.Printf("code-radar %s (commit: %s, built: %s)\n", version, commit, date)
	fmt.Println()
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func renderDirect(commitHash string, jsonOut, shortOut bool, cfg Config) {
	if err := verifyCommit(commitHash); err != nil {
		fmt.Fprintln(os.Stderr, errorIcon.Render(" Error: ")+err.Error())
		os.Exit(1)
	}

	commitInfo, err := getCommitInfo(commitHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, errorIcon.Render(" Error: ")+err.Error())
		os.Exit(1)
	}

	branches, err := findBranchesWithCommit(commitHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, errorIcon.Render(" Error: ")+err.Error())
		os.Exit(1)
	}

	parts := strings.Split(commitInfo, "|")
	result := scanResult{
		Hash:    parts[0],
		Message: safeIndex(parts, 1),
		Author:  safeIndex(parts, 2),
		Date:    safeIndex(parts, 3),
	}

	for _, b := range branches {
		env := matchEnvironment(b, cfg)
		result.Branches = append(result.Branches, branchInfo{
			Name: b,
			Env:  env,
		})
	}

	envMap := categorizeBranches(branches, cfg)
	for _, env := range cfg.Environments {
		if bs, ok := envMap[env.Name]; ok {
			result.Pipeline = append(result.Pipeline, pipelineEnv{
				Name:     env.Name,
				Branches: bs,
			})
		}
	}
	if bs, ok := envMap["OTHER"]; ok {
		result.Pipeline = append(result.Pipeline, pipelineEnv{
			Name:     "OTHER",
			Branches: bs,
		})
	}

	if jsonOut {
		fmt.Print(renderJSON(result))
	} else if shortOut {
		fmt.Print(renderShort(result))
	} else {
		fmt.Print(renderFull(result, cfg))
	}
}
