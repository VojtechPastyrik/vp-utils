package clean_branches

import (
	"os/exec"
	"strings"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/git"
	"github.com/VojtechPastyrik/vp-utils/pkg/logger"
	"github.com/spf13/cobra"
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
	parent_cmd.Cmd.AddCommand(Cmd)
}

func cleanBranches() {
	logger.Info("starting branch cleanup...")
	removedMerged := removeMergedBranches()
	logger.Infof("removed %d merged branches.\n", removedMerged)

	removedOld := removeOldBranches()
	logger.Infof("removed %d old branches.\n", removedOld)

	removedNoUpstream := removeBranchesWithoutUpstream()
	logger.Infof("removed %d branches without upstream.\n", removedNoUpstream)

	logger.Info("branch cleanup completed.")
}

func removeMergedBranches() int {
	var deletedCount int
	out, err := exec.Command("git", "branch", "--merged", "main").Output()
	if err != nil {
		logger.Fatalf("error when getting merged branches: %v", err)
	}
	branches := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" || branch == "main" || branch == "master" {
			continue
		}
		cmd := exec.Command("git", "branch", "-d", branch)
		if err := cmd.Run(); err != nil {
			logger.Infof("failed to delete merged branch '%s': %v", branch, err)
		} else {
			logger.Infof("successfully deleted merged branch: '%s'", branch)
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
		logger.Fatalf("error when retrieving branches for old branch cleanup: %v", err)
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
			logger.Infof("failed to parse date for branch '%s': %v", branch, err)
			continue
		}

		// Check if branch is older than 6 months
		if commitDate.Before(sixMonthsAgo) {
			cmd := exec.Command("git", "branch", "-d", branch)
			if err := cmd.Run(); err != nil {
				logger.Infof("failed to delete old branch '%s': %v", branch, err)
			} else {
				logger.Infof("successfully deleted old branch: '%s'", branch)
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
		logger.Fatalf("error when getting branches to check for upstream: %v", err)
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
				logger.Infof("failed to delete branch without upstream '%s': %v", branch, err)
			} else {
				logger.Infof("successfully deleted branch without upstream: '%s'", branch)
				deletedCount++
			}
		}
	}
	return deletedCount
}
