package release

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/VojtechPastyrik/vp-utils/cmd/root"
	"github.com/VojtechPastyrik/vp-utils/pkg/logger"
	git "github.com/go-git/go-git/v5"
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
	version := getVersionFromFile()
	version = strings.TrimSuffix(version, "-dev")

	writeNewVersionFile(version)

	splitVersion := strings.Split(strings.TrimPrefix(version, "v"), ".")
	minorVersion, _ := strconv.Atoi(splitVersion[1])
	minorVersion++
	newVersion := fmt.Sprintf("v%s.%d.%s-dev", splitVersion[0], minorVersion, splitVersion[2])

	// Load the git repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		logger.Fatalf("error opening git repository: %v\n", err)

	}

	// Check if the current branch is 'main'
	branch, err := repo.Head()
	if err != nil {
		logger.Fatalf("error getting current branch: %v\n", err)
	}
	if branch.Name().Short() != "main" {
		logger.Fatalf("current branch is not 'main', but '%s'\n", branch.Name().Short())
	}

	// Get actual state of the files in repository
	w, err := repo.Worktree()
	if err != nil {
		logger.Fatalf("error getting worktree: %v\n", err)
	}
	// Add all changes to the staging area
	_, err = w.Add(".")
	if err != nil {
		logger.Fatalf("error adding file: %v\n", err)
	}
	// Commit the changes
	commit, err := w.Commit("VERSION: "+version, &git.CommitOptions{})
	if err != nil {
		logger.Fatalf("error committing changes: %v\n", err)
	}
	// Create a new tag with the new version
	_, err = repo.CreateTag(version, commit, nil)
	if err != nil {
		logger.Fatalf("error creating tag: %v\n", err)
	}

	// Write the new version file
	writeNewVersionFile(newVersion)
	_, err = w.Add(".")
	if err != nil {
		logger.Fatalf("error adding file: %v\n", err)
	}
	// Commit the new version file
	commit, err = w.Commit("VERSION: "+newVersion, &git.CommitOptions{})
	if err != nil {
		logger.Fatalf("error committing changes: %v\n", err)
	}
	// Push the new version file to the remote repository
	cmd := exec.Command("git", "push", "origin", branch.Name().Short(), "--tags")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("error pushing changes: %v\nOutput: %s", err, string(output))
	}
	logger.Infof("successfully released version: %s\n", version)
	logger.Infof("successfully updated version to: %s\n", newVersion)
}

func writeNewVersionFile(version string) {
	fileContent := fmt.Sprintf(`package version

var Version string = "%s"`, version)

	err := os.WriteFile("version/version.go", []byte(fileContent), 0644)
	if err != nil {
		logger.Fatalf("error writing new version file: %v\n", err)
	}

}

func getVersionFromFile() string {
	data, err := os.ReadFile("version/version.go")
	if err != nil {
		logger.Fatalf("error reading version file: %v\n", err)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "var Version string") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), `"`)
			}
		}
	}
	logger.Fatal("version not found in version.go")
	return ""
}
