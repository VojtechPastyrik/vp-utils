package hosts

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/ssh"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	configPath string
	filter     string
)

type hostEntry struct {
	Name     string
	HostName string
	User     string
	Port     string
	Source   string
}

var Cmd = &cobra.Command{
	Use:   "hosts",
	Short: "List SSH hosts from ~/.ssh/config (including Includes)",
	Run: func(cmd *cobra.Command, args []string) {
		path := configPath
		if path == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				logger.Fatalf("cannot get home directory: %v", err)
			}
			path = filepath.Join(home, ".ssh", "config")
		}

		entries, err := parseConfig(path)
		if err != nil {
			logger.Fatalf("error parsing SSH config: %v", err)
		}

		if filter != "" {
			filtered := make([]hostEntry, 0)
			for _, e := range entries {
				if strings.Contains(strings.ToLower(e.Name), strings.ToLower(filter)) ||
					strings.Contains(strings.ToLower(e.HostName), strings.ToLower(filter)) {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}

		if len(entries) == 0 {
			fmt.Println("No SSH hosts found")
			return
		}

		fmt.Printf("%-30s %-30s %-10s %-6s %s\n", "HOST", "HOSTNAME", "USER", "PORT", "SOURCE")
		fmt.Printf("%-30s %-30s %-10s %-6s %s\n", "----", "--------", "----", "----", "------")
		for _, e := range entries {
			port := e.Port
			if port == "" {
				port = "22"
			}
			fmt.Printf("%-30s %-30s %-10s %-6s %s\n", e.Name, e.HostName, e.User, port, e.Source)
		}
		fmt.Printf("\nTotal: %d hosts\n", len(entries))
	},
}

func init() {
	Cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to SSH config file (default: ~/.ssh/config)")
	Cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter hosts by name or hostname")
	parent_cmd.Cmd.AddCommand(Cmd)
}

func parseConfig(path string) ([]hostEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dir := filepath.Dir(path)
	var entries []hostEntry
	var current *hostEntry
	// Track inherited defaults (Host * or wildcard patterns with only settings)
	var defaults hostEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value := parseLine(line)
		switch strings.ToLower(key) {
		case "include":
			included, err := resolveInclude(dir, value)
			if err != nil {
				logger.Warnf("cannot resolve Include %s: %v", value, err)
				continue
			}
			for _, incPath := range included {
				info, err := os.Stat(incPath)
				if err != nil || info.IsDir() {
					continue
				}
				incEntries, err := parseConfig(incPath)
				if err != nil {
					logger.Warnf("error parsing %s: %v", incPath, err)
					continue
				}
				entries = append(entries, incEntries...)
			}
		case "host":
			// Save previous entry
			if current != nil && !isWildcard(current.Name) {
				applyDefaults(current, defaults)
				entries = append(entries, *current)
			}
			if isWildcard(value) {
				current = nil
				// This is a wildcard block, update defaults from it
				defaults.Name = value
			} else {
				current = &hostEntry{Name: value, Source: filepath.Base(path)}
			}
		case "hostname":
			if current != nil {
				current.HostName = value
			}
		case "user":
			if current != nil {
				current.User = value
			} else {
				defaults.User = value
			}
		case "port":
			if current != nil {
				current.Port = value
			} else {
				defaults.Port = value
			}
		}
	}

	// Don't forget the last entry
	if current != nil && !isWildcard(current.Name) {
		applyDefaults(current, defaults)
		entries = append(entries, *current)
	}

	return entries, scanner.Err()
}

func parseLine(line string) (string, string) {
	// Handle both "Key Value" and "Key=Value" formats
	if idx := strings.Index(line, "="); idx != -1 {
		return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:])
	}
	parts := strings.SplitN(line, " ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return line, ""
}

func resolveInclude(baseDir, pattern string) ([]string, error) {
	if strings.HasPrefix(pattern, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		pattern = filepath.Join(home, pattern[1:])
	} else if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(baseDir, pattern)
	}
	return filepath.Glob(pattern)
}

func isWildcard(name string) bool {
	return strings.Contains(name, "*") || strings.Contains(name, "?")
}

func applyDefaults(entry *hostEntry, defaults hostEntry) {
	if entry.User == "" && defaults.User != "" {
		entry.User = defaults.User
	}
	if entry.Port == "" && defaults.Port != "" {
		entry.Port = defaults.Port
	}
}
