package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func main() {
	var commitHash string

	if len(os.Args) > 1 {
		commitHash = os.Args[1]
	} else {
		fmt.Print("Enter commit hash (full or short): ")
		fmt.Scanln(&commitHash)
	}

	commitHash = strings.TrimSpace(commitHash)
	if commitHash == "" {
		fmt.Fprintln(os.Stderr, "Error: commit hash is required")
		os.Exit(1)
	}

	// Verify commit exists
	if err := verifyCommit(commitHash); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Get commit info
	commitInfo, err := getCommitInfo(commitHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting commit info: %v\n", err)
		os.Exit(1)
	}

	// Find branches containing this commit
	branches, err := findBranchesWithCommit(commitHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding branches: %v\n", err)
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
	// Use --contains to find all branches that contain the commit
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
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    CODE COMMIT TRACKER                       ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ Commit:  %-50s ║\n", truncateString(fullHash, 50))
	if message != "" {
		fmt.Printf("║ Message: %-50s ║\n", truncateString(message, 50))
	}
	if author != "" {
		fmt.Printf("║ Author:  %-50s ║\n", truncateString(author, 50))
	}
	if date != "" {
		fmt.Printf("║ Date:    %-50s ║\n", truncateString(date, 50))
	}
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	if len(branches) == 0 {
		fmt.Println("║ No branches contain this commit                              ║")
	} else {
		fmt.Printf("║ Found in %d branch(es):%s║\n", len(branches), strings.Repeat(" ", 41-len(fmt.Sprintf("%d", len(branches)))))
		fmt.Println("╠══════════════════════════════════════════════════════════════╣")

		for _, branch := range branches {
			env := categorizeSingleBranch(branch)
			fmt.Printf("║  %s %-52s ║\n", env, branch)
		}
	}

	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Summary
	if len(branches) > 0 {
		fmt.Println("Deployment Summary:")
		fmt.Println("-------------------")
		envs := make(map[string][]string)
		for _, b := range branches {
			env := categorizeSingleBranch(b)
			envs[env] = append(envs[env], b)
		}
		envOrder := []string{"[DEV]", "[QA]", "[UAT]", "[STAGING]", "[MASTER]", "[PROD]", "[OTHER]"}
		for _, env := range envOrder {
			if bs, ok := envs[env]; ok {
				fmt.Printf("  %s %s\n", env, strings.Join(bs, ", "))
			}
		}
		fmt.Println()
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
		env := categorizeSingleBranch(b)
		result[env] = append(result[env], b)
	}
	return result
}

func categorizeSingleBranch(branch string) string {
	lower := strings.ToLower(branch)
	switch {
	case lower == "master" || lower == "main":
		return "[MASTER]"
	case lower == "prod" || lower == "production":
		return "[PROD]"
	case strings.Contains(lower, "staging") || strings.Contains(lower, "stg"):
		return "[STAGING]"
	case strings.Contains(lower, "uat"):
		return "[UAT]"
	case strings.Contains(lower, "qa") || strings.Contains(lower, "test"):
		return "[QA]"
	case lower == "develop" || lower == "dev":
		return "[DEV]"
	default:
		return "[OTHER]"
	}
}
