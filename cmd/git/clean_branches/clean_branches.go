package clean_branches

import (
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/git"
	"github.com/spf13/cobra"
	"log"
	"os/exec"
	"strconv"
	"strings"
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
	removeMergedBranches()
	log.Println("Old and merged branches removed successfully.")
	removeOldBranches()
	log.Println("Old branches removed successfully.")
	removeBranchesWithoutUpstream()
	log.Println("Branches without upstream removed successfully.")
}

func removeMergedBranches() {
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
			log.Printf("Failed to delete branch %s: %v", branch, err)
		} else {
			log.Printf("Deleted merged branch: %s", branch)
		}
	}
}

func removeOldBranches() {
	out, err := exec.Command("git", "for-each-ref", "--sort=-committerdate", "--format=%(refname:short) %(committerdate:relative)", "refs/heads/").Output()
	if err != nil {
		log.Fatalf("Error when retrieving branches: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		branch := parts[0]
		age := parts[1]

		// Check if branch is older than 6 months
		if strings.Contains(age, "month") {
			months := strings.Split(age, " ")[0]
			if monthsInt, err := strconv.Atoi(months); err == nil && monthsInt >= 6 {
				cmd := exec.Command("git", "branch", "-d", branch)
				if err := cmd.Run(); err != nil {
					log.Printf("Failed to delete branch %s: %v", branch, err)
				} else {
					log.Printf("Deleted old branch: %s", branch)
				}
			}
		}
	}
}

func removeBranchesWithoutUpstream() {
	out, err := exec.Command("git", "branch", "-vv").Output()
	if err != nil {
		log.Fatalf("Chyba při získávání větví: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.Contains(line, ": gone") {
			branch := strings.Fields(line)[0]
			cmd := exec.Command("git", "branch", "-d", branch)
			if err := cmd.Run(); err != nil {
				log.Printf("Chyba při mazání větve %s: %v", branch, err)
			} else {
				log.Printf("Větev bez upstreamu %s byla úspěšně smazána.", branch)
			}
		}
	}
}
