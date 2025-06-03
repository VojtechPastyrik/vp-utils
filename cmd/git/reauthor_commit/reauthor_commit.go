package reauthor_commit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/git"
	"github.com/spf13/cobra"
)

var (
	FlagUsername   string
	FlagEmail      string
	FlagCommitHash string
)

var Cmd = &cobra.Command{
	Use:     "reassign-commit",
	Short:   "Change Git commit author",
	Aliases: []string{"rc", "reauthor"},
	Run: func(cmd *cobra.Command, args []string) {
		if FlagCommitHash != "" {
			changeSpecificCommitAuthor(FlagCommitHash, FlagUsername, FlagEmail)
		} else {
			changeLastCommitAuthor(FlagUsername, FlagEmail)
		}
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagUsername, "username", "u", "", "New author name for the commit")
	Cmd.MarkFlagRequired("username")
	Cmd.Flags().StringVarP(&FlagEmail, "email", "e", "", "New author email for the commit")
	Cmd.MarkFlagRequired("email")
	Cmd.Flags().StringVarP(&FlagCommitHash, "commit", "c", "", "Hash of the commit to change (defaults to last commit if not specified)")
}

func changeLastCommitAuthor(username, email string) {
	cmd := exec.Command("git", "commit", "--amend", "--author", fmt.Sprintf("%s <%s>", username, email), "--no-edit")
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Error changing last commit author: %v\n%s\n", err, output)
		return
	}
	fmt.Println("Last commit author successfully changed")
}

func changeSpecificCommitAuthor(commitHash, username, email string) {
	// Check if commit hash exists
	checkCmd := exec.Command("git", "cat-file", "-t", commitHash)
	if err := checkCmd.Run(); err != nil {
		fmt.Printf("Commit with hash '%s' not found\n", commitHash)
		return
	}

	// Create temporary script
	scriptContent := fmt.Sprintf(`
if [ "$GIT_COMMIT" = "%s" ]
then
    export GIT_AUTHOR_NAME="%s"
    export GIT_AUTHOR_EMAIL="%s"
    export GIT_COMMITTER_NAME="%s"
    export GIT_COMMITTER_EMAIL="%s"
fi
`, commitHash, username, email, username, email)

	tmpFile, err := createTempScript(scriptContent)
	if err != nil {
		fmt.Printf("Error creating temporary script: %v\n", err)
		return
	}
	defer cleanupTempScript(tmpFile)

	// Run filter-branch
	rebaseCmd := exec.Command("git", "filter-branch", "--env-filter", fmt.Sprintf("source %s", tmpFile), "--force", commitHash+"^..HEAD")
	rebaseOutput, err := rebaseCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error changing commit author: %v\n%s\n", err, rebaseOutput)
		return
	}

	fmt.Printf("Author of commit %s successfully changed\n", commitHash)
	fmt.Println("Warning: This operation has rewritten Git history. If you've already published these commits, you must use 'git push --force'")
}

func createTempScript(content string) (string, error) {
	// Create temporary file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "git-reauthor-script.sh")

	// Write content to file
	err := os.WriteFile(tmpFile, []byte(content), 0755)
	if err != nil {
		return "", err
	}

	return tmpFile, nil
}

func cleanupTempScript(path string) {
	// Remove temporary file
	os.Remove(path)
}
