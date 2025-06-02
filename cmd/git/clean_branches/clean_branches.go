package clean_branches

import (
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/git"
	"github.com/spf13/cobra"
	"log"
	"os/exec"
	"strings"
	"time"
)

var Cmd = &cobra.Command{
	Use:     "clean-branches",
	Short:   "Remove old and merged branches from git on local",
	Aliases: []string{"cb"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cleanBranches()
	},
}

func init() {
	parent_cmd.Cmd.Add("clean-branches", Cmd) // Assuming parent_cmd.Cmd.Add is the correct way to add
}

func cleanBranches() {
	log.Println("Starting branch cleanup...")
	removedMerged := removeMergedBranches()
	log.Printf("Removed %d merged branches.\n", removedMerged)

	removedOld := removeOldBranches()
	log.Printf("Removed %d old branches.\n", removedOld)

	removedNoUpstream := removeBranchesWithoutUpstream()
	log.Printf("Removed %d branches without upstream.\n", removedNoUpstream)

	log.Println("Branch cleanup completed.")
}

func removeMergedBranches() int {
	var deletedCount int
	out, err := exec.Command("git", "branch", "--merged", "main").Output()
	if err != nil {
		log.Fatalf("Error when getting merged branches: %v", err)
	}
	branches := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" || branch == "main" || branch == "master" {
			continue
		}
		cmd := exec.Command("git", "branch", "-d", branch)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to delete merged branch '%s': %v", branch, err)
		} else {
			log.Printf("Successfully deleted merged branch: '%s'", branch)
			deletedCount++
		}
	}
	return deletedCount
}

func removeOldBranches() int {
	var deletedCount int
	// Get branches with their last commit date in ISO 8601 format
	out, err := exec.Command("git", "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)|%(committerdate:iso8601)", "refs/heads/").Output()
	if err != nil {
		log.Fatalf("Error when retrieving branches for old branch cleanup: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	sixMonthsAgo := time.Now().AddDate(0, -6, 0) // Calculate date 6 months ago

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		branch := parts[0]
		dateStr := parts[1]

		// Skip main and master branches
		if branch == "main" || branch == "master" {
			continue
		}

		commitDate, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			log.Printf("Failed to parse date for branch '%s': %v", branch, err)
			continue
		}

		// Check if branch is older than 6 months
		if commitDate.Before(sixMonthsAgo) {
			cmd := exec.Command("git", "branch", "-d", branch)
			if err := cmd.Run(); err != nil {
				log.Printf("Failed to delete old branch '%s': %v", branch, err)
			} else {
				log.Printf("Successfully deleted old branch: '%s'", branch)
				deletedCount++
			}
		}
	}
	return deletedCount
}

func removeBranchesWithoutUpstream() int {
	var deletedCount int
	out, err := exec.Command("git", "branch", "-vv").Output()
	if err != nil {
		log.Fatalf("Error when getting branches to check for upstream: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		// Example output for a branch without upstream (or gone):
		//   feature/my-branch    abcdefg [origin/feature/my-branch: gone] Some commit message
		//   another-branch       abcdefg
		if strings.Contains(line, ": gone") {
			// Extract branch name (first field before whitespace)
			branch := strings.Fields(line)[0]
			branch = strings.TrimPrefix(branch, "* ") // Remove asterisk for current branch if present

			// Skip main and master branches
			if branch == "main" || branch == "master" {
				continue
			}

			cmd := exec.Command("git", "branch", "-d", branch)
			if err := cmd.Run(); err != nil {
				log.Printf("Failed to delete branch without upstream '%s': %v", branch, err)
			} else {
				log.Printf("Successfully deleted branch without upstream: '%s'", branch)
				deletedCount++
			}
		}
	}
	return deletedCount
}
