package track_time

import (
	"bytes"
	"encoding/json"
	"fmt"
	parent_cmd "github.com/VojtechPastyrik/vp-utils/cmd/youtrack"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var (
	FlagToken   string
	FlagUrl     string
	FlagIssueId string
	FlagNote    string
	FlagDate    string
	FlagMinutes int
	FlagHours   int
)

var Cmd = &cobra.Command{
	Use:     "track-time",
	Short:   "Track time spent on tasks in YouTrack",
	Long:    `Track time spent on tasks in YouTrack.`,
	Example: `vp-utils youtrack track-time --token <token> --url <url> --issue-id <issue-id> --note "Worked on feature X" --date "2023-10-01" --minutes 30`,
	Args:    cobra.NoArgs,
	Aliases: []string{"tt"},
	Run: func(cmd *cobra.Command, args []string) {
		trackTime(FlagToken, FlagUrl, FlagIssueId, FlagNote, FlagDate, FlagMinutes, FlagHours)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagToken,
		"token",
		"t",
		"",
		"Personal access token for YouTrack",
	)
	Cmd.MarkFlagRequired("token")
	Cmd.Flags().StringVarP(
		&FlagUrl,
		"url",
		"u",
		"",
		"URL of the YouTrack instance",
	)
	Cmd.MarkFlagRequired("url")
	Cmd.Flags().StringVarP(
		&FlagIssueId,
		"issue-id",
		"i",
		"",
		"ID of the issue to track time for",
	)
	Cmd.MarkFlagRequired("issue-id")
	Cmd.Flags().StringVarP(
		&FlagNote,
		"note",
		"n",
		"",
		"Note to add to the time tracking entry",
	)
	Cmd.Flags().StringVarP(
		&FlagDate,
		"date",
		"d",
		"",
		"Date for the time tracking entry (format: YYYY-MM-DD)",
	)
	Cmd.Flags().IntVarP(
		&FlagMinutes,
		"minutes",
		"m",
		0,
		"Number of minutes spent on the task",
	)
	Cmd.Flags().IntVarP(
		&FlagHours,
		"hours",
		"H",
		0,
		"Number of hours spent on the task",
	)
}

type Duration struct {
	Minutes int `json:"minutes"`
}

type WorkItem struct {
	Text     string   `json:"text"`
	Duration Duration `json:"duration"`
	Date     int64    `json:"date"`
}

func trackTime(token, ytUrl, issueId, note, date string, minutes int, hours int) {
	if hours == 0 && minutes == 0 {
		log.Fatalln("You must specify either minutes or hours to track time")
	}
	if hours > 0 {
		minutes += hours * 60
	}
	var dateUnixMilli int64
	if date != "" {
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			log.Fatalf("Invalid date format: %v. Use YYYY-MM-DD format", err)
		}
		dateUnixMilli = parsedDate.UnixMilli()
	} else {
		dateUnixMilli = time.Now().UnixMilli()
	}
	url := fmt.Sprintf("%s/api/issues/%s/timeTracking/workItems", ytUrl, issueId)
	work := WorkItem{
		Text: note,
		Duration: Duration{
			Minutes: minutes,
		},
		Date: dateUnixMilli,
	}

	workJson, err := json.Marshal(work)
	if err != nil {
		log.Fatalln("Error marshalling JSON:", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(workJson))
	if err != nil {
		log.Fatalln("Error creating request:", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error making request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Failed to track time in YouTrack. API returned: %s (HTTP %d)", resp.Status, resp.StatusCode)
	} else {
		log.Println("Time tracked successfully in YouTrack")
	}
}
