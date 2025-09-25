package release

import (
	"fmt"
	"os/exec"

	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	version "github.com/VojtechPastyrik/vp-utils/version"
	git "github.com/go-git/go-git/v5"

	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "release",
	Short:   "Release vp-utils",
	Long:    "Release vp-utils provides command to release vp-utils.",
	Example: `vp-utils release -`,
	Run: func(cmd *cobra.Command, args []string) {
		releaseProject()
	},
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}

func releaseProject() {
	version := version.Version
	version = strings.TrimSuffix(version, "-dev")

	writeNewVersionFile(version)

	splitVersion := strings.Split(strings.TrimPrefix(version, "v"), ".")
	minorVersion, _ := strconv.Atoi(splitVersion[1])
	minorVersion++
	newVersion := fmt.Sprintf("v%s.%d.%s-dev", splitVersion[0], minorVersion, splitVersion[2])

	// Load the git repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Fatalf("Error opening git repository: %v\n", err)

	}

	// Check if the current branch is 'main'
	branch, err := repo.Head()
	if err != nil {
		log.Fatalf("Error getting current branch: %v\n", err)
	}
	if branch.Name().Short() != "main" {
		log.Fatalf("Current branch is not 'main', but '%s'\n", branch.Name().Short())
	}

	// Get actual state of the files in repository
	w, err := repo.Worktree()
	if err != nil {
		log.Fatalf("Error getting worktree: %v\n", err)
	}
	// Add all changes to the staging area
	_, err = w.Add(".")
	if err != nil {
		log.Fatalf("Error adding file: %v\n", err)
	}
	// Commit the changes
	commit, err := w.Commit("VERSION: "+version, &git.CommitOptions{})
	if err != nil {
		log.Fatalf("Error committing changes: %v\n", err)
	}
	// Create a new tag with the new version
	_, err = repo.CreateTag(version, commit, nil)
	if err != nil {
		log.Fatalf("Error creating tag: %v\n", err)
	}

	// Write the new version file
	writeNewVersionFile(newVersion)
	_, err = w.Add(".")
	if err != nil {
		log.Fatalf("Error adding file: %v\n", err)
	}
	// Commit the new version file
	commit, err = w.Commit("VERSION: "+newVersion, &git.CommitOptions{})
	if err != nil {
		log.Fatalf("Error committing changes: %v\n", err)
	}
	// Push the new version file to the remote repository
	cmd := exec.Command("git", "push", "origin", branch.Name().Short(), "--tags")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error pushing changes: %v\nOutput: %s", err, string(output))
	}
	log.Printf("Successfully released version: %s\n", version)
	log.Printf("Successfully updated version to: %s\n", newVersion)
}

func writeNewVersionFile(version string) {
	fileContent := fmt.Sprintf(`package version

var Version string = "%s"`, version)

	err := os.WriteFile("version/version.go", []byte(fileContent), 0644)
	if err != nil {
		log.Fatalf("Error writing new version file: %v\n", err)
	}

}
