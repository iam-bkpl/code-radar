package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

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
			Foreground(lipgloss.Color("#8BE9FD"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	hashStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#50FA7B"))

	branchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			PaddingLeft(2)

	// Environment tag styles
	devStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#6272A4")).
			Padding(0, 1)

	qaStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#282A36")).
			Background(lipgloss.Color("#F1FA8C")).
			Padding(0, 1)

	uatStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#282A36")).
			Background(lipgloss.Color("#FFB86C")).
			Padding(0, 1)

	stagingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#282A36")).
			Background(lipgloss.Color("#8BE9FD")).
			Padding(0, 1)

	masterStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#BD93F9")).
			Padding(0, 1)

	prodStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#FF5555")).
			Padding(0, 1)

	otherStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#6272A4")).
			Padding(0, 1)

	successIcon = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	errorIcon = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	checkMark = successIcon.Render(" ✓ ")
	crossMark = errorIcon.Render(" ✗ ")
	arrow     = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Render(" → ")
	dot       = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(" • ")
)

func main() {
	// Handle flags
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help", "-h":
			printHelp()
			return
		case "--version", "-v":
			printVersion()
			return
		}
	}

	var commitHash string

	if len(os.Args) > 1 {
		commitHash = os.Args[1]
	} else {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("  📡 Code Radar"))
		fmt.Println()
		fmt.Print("  Enter commit hash: ")
		fmt.Scanln(&commitHash)
	}

	commitHash = strings.TrimSpace(commitHash)
	if commitHash == "" {
		fmt.Fprintln(os.Stderr, errorIcon.Render(" Error: ")+"commit hash is required")
		os.Exit(1)
	}

	// Verify commit exists
	if err := verifyCommit(commitHash); err != nil {
		fmt.Println()
		fmt.Println(errorIcon.Render(" Error: ") + err.Error())
		fmt.Println()
		os.Exit(1)
	}

	// Get commit info
	commitInfo, err := getCommitInfo(commitHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, errorIcon.Render(" Error getting commit info: ")+err.Error())
		os.Exit(1)
	}

	// Find branches containing this commit
	branches, err := findBranchesWithCommit(commitHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, errorIcon.Render(" Error finding branches: ")+err.Error())
		os.Exit(1)
	}

	// Display results
	printResults(commitHash, commitInfo, branches)
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
	cmd := exec.Command("git", "branch", "--contains", hash, "--format=%(refname:short)")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}

	sort.Strings(branches)
	return branches, nil
}

func printResults(hash string, info string, branches []string) {
	parts := strings.Split(info, "|")
	fullHash := parts[0]
	message := ""
	author := ""
	date := ""
	if len(parts) > 1 {
		message = parts[1]
	}
	if len(parts) > 2 {
		author = parts[2]
	}
	if len(parts) > 3 {
		date = parts[3]
	}

	fmt.Println()

	// Title
	title := titleStyle.Render("  📡  CODE RADAR  ")
	fmt.Println(title)
	fmt.Println()

	// Commit info section
	fmt.Println(headerStyle.Render("  ┌─ COMMIT DETAILS"))
	fmt.Println("  │")
	fmt.Printf("  %s %s\n", labelStyle.Render("Hash:"), hashStyle.Render(fullHash[:12]))
	if message != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Message:"), valueStyle.Render(truncateString(message, 55)))
	}
	if author != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Author:"), valueStyle.Render(author))
	}
	if date != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Date:"), valueStyle.Render(truncateString(date, 40)))
	}
	fmt.Println("  │")
	fmt.Println()

	// Branches section
	if len(branches) == 0 {
		fmt.Println(headerStyle.Render("  ┌─ BRANCHES"))
		fmt.Println("  │")
		fmt.Printf("  %s %s\n", crossMark, errorIcon.Render("No branches contain this commit"))
		fmt.Println("  │")
	} else {
		fmt.Println(headerStyle.Render(fmt.Sprintf("  ┌─ BRANCHES (%d found)", len(branches))))
		fmt.Println("  │")

		for i, branch := range branches {
			envTag, style := getBranchStyle(branch)
			connector := "├"
			if i == len(branches)-1 {
				connector = "└"
			}

			fmt.Printf("  %s %s %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(connector+"─"),
				arrow,
				branchStyle.Render(style.Render(envTag)+" "+branch),
			)
		}
		fmt.Println("  │")
	}

	// Summary section
	if len(branches) > 0 {
		fmt.Println(headerStyle.Render("  ┌─ DEPLOYMENT PIPELINE"))
		fmt.Println("  │")
		printPipeline(branches)
		fmt.Println("  │")
	}

	fmt.Println()
}

func printPipeline(branches []string) {
	envs := categorizeBranches(branches)

	// Define pipeline order
	pipeline := []struct {
		name  string
		style lipgloss.Style
	}{
		{"DEV", devStyle},
		{"QA", qaStyle},
		{"UAT", uatStyle},
		{"STAGING", stagingStyle},
		{"MASTER", masterStyle},
		{"PROD", prodStyle},
	}

	for i, p := range pipeline {
		if bs, ok := envs[p.name]; ok {
			connector := "├"
			if i == len(pipeline)-1 {
				connector = "└"
			}

			fmt.Printf("  %s %s %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(connector+"─"),
				p.style.Render(fmt.Sprintf(" %s ", p.name)),
				valueStyle.Render(strings.Join(bs, ", ")),
			)
		}
	}

	// Other branches
	if bs, ok := envs["OTHER"]; ok {
		fmt.Printf("  %s %s %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render("└─"),
			otherStyle.Render(" OTHER"),
			valueStyle.Render(strings.Join(bs, ", ")),
		)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func categorizeBranches(branches []string) map[string][]string {
	result := make(map[string][]string)
	for _, b := range branches {
		env := categorizeBranchName(b)
		result[env] = append(result[env], b)
	}
	return result
}

func categorizeBranchName(branch string) string {
	lower := strings.ToLower(branch)
	switch {
	case lower == "master" || lower == "main":
		return "MASTER"
	case lower == "prod" || lower == "production":
		return "PROD"
	case strings.Contains(lower, "staging") || strings.Contains(lower, "stg"):
		return "STAGING"
	case strings.Contains(lower, "uat"):
		return "UAT"
	case strings.Contains(lower, "qa") || strings.Contains(lower, "test"):
		return "QA"
	case lower == "develop" || lower == "dev":
		return "DEV"
	default:
		return "OTHER"
	}
}

func getBranchStyle(branch string) (string, lipgloss.Style) {
	env := categorizeBranchName(branch)
	switch env {
	case "DEV":
		return "DEV", devStyle
	case "QA":
		return "QA", qaStyle
	case "UAT":
		return "UAT", uatStyle
	case "STAGING":
		return "STAGING", stagingStyle
	case "MASTER":
		return "MASTER", masterStyle
	case "PROD":
		return "PROD", prodStyle
	default:
		return "OTHER", otherStyle
	}
}

func printHelp() {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("  📡 Code Radar"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render("  Track where your git commits have been deployed"))
	fmt.Println()

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
	commandStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	flagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))

	fmt.Println(helpStyle.Render("  Usage:"))
	fmt.Println()
	fmt.Printf("    %s %s\n", commandStyle.Render("code-radar"), flagStyle.Render("<commit-hash>"))
	fmt.Println(helpStyle.Render("      Track a specific commit across branches"))
	fmt.Println()
	fmt.Printf("    %s\n", commandStyle.Render("code-radar"))
	fmt.Println(helpStyle.Render("      Interactive mode - prompts for commit hash"))
	fmt.Println()

	fmt.Println(helpStyle.Render("  Flags:"))
	fmt.Println()
	fmt.Printf("    %s, %s\n", flagStyle.Render("--help, -h"), helpStyle.Render("Show this help message"))
	fmt.Printf("    %s, %s\n", flagStyle.Render("--version, -v"), helpStyle.Render("Show version information"))
	fmt.Println()

	fmt.Println(helpStyle.Render("  Examples:"))
	fmt.Println()
	fmt.Println(helpStyle.Render("    code-radar a1b2c3d"))
	fmt.Println(helpStyle.Render("    code-radar abc1234def5678"))
	fmt.Println(helpStyle.Render("    code-radar"))
	fmt.Println()
}

func printVersion() {
	fmt.Println()
	fmt.Printf("code-radar %s (commit: %s, built: %s)\n", version, commit, date)
	fmt.Println()
}
