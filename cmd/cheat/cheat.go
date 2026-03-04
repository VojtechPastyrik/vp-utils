package cheat

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

const (
	baseURL  = "https://raw.githubusercontent.com/tldr-pages/tldr/main/pages"
	cacheTTL = 7 * 24 * time.Hour
)

var (
	platform string
	noCache  bool
)

var Cmd = &cobra.Command{
	Use:   "cheat <command>",
	Short: "Quick command cheatsheets powered by tldr-pages",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		command := args[0]
		plat := detectPlatform()

		// Try cache first
		if !noCache {
			if content, ok := readCache(command, plat); ok {
				renderPage(content)
				return
			}
		}

		// Fetch from network, try platform-specific then common
		for _, p := range []string{plat, "common"} {
			url := fmt.Sprintf("%s/%s/%s.md", baseURL, p, command)
			content, err := fetchPage(url)
			if err == nil {
				writeCache(command, p, content)
				renderPage(content)
				return
			}
		}

		logger.Fatalf("no cheatsheet found for %q", command)
	},
}

func init() {
	Cmd.Flags().StringVar(&platform, "platform", "", "Platform: linux, osx, common (default: auto-detect)")
	Cmd.Flags().BoolVar(&noCache, "no-cache", false, "Force re-fetch from network")
	root.RootCmd.AddCommand(Cmd)
}

func detectPlatform() string {
	if platform != "" {
		return platform
	}
	switch runtime.GOOS {
	case "darwin":
		return "osx"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return "common"
	}
}

func cacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "vpd", "tldr")
}

func cachePath(command, plat string) string {
	return filepath.Join(cacheDir(), plat, command+".md")
}

func readCache(command, plat string) (string, bool) {
	// Try platform-specific, then common
	for _, p := range []string{plat, "common"} {
		path := cachePath(command, p)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if time.Since(info.ModTime()) > cacheTTL {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		return string(data), true
	}
	return "", false
}

func writeCache(command, plat, content string) {
	path := cachePath(command, plat)
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(content), 0644)
}

func fetchPage(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func renderPage(content string) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "# "):
			fmt.Printf("\033[1;36m%s\033[0m\n", line[2:])
		case strings.HasPrefix(line, "> "):
			fmt.Printf("\033[90m  %s\033[0m\n", line[2:])
		case strings.HasPrefix(line, "- "):
			fmt.Printf("\n\033[32m  %s\033[0m\n", line[2:])
		case strings.HasPrefix(line, "`") && strings.HasSuffix(line, "`"):
			fmt.Printf("\033[1;33m    %s\033[0m\n", line[1:len(line)-1])
		case line == "":
			// skip blank lines
		default:
			fmt.Println(line)
		}
	}
}
