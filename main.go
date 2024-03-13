package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"

	"jira-helper/services/jira"
)

func main() {
	authToken := os.Getenv("JIRA_AUTH_TOKEN")
	meetingsJiraIssue := os.Getenv("MEETINGS_JIRA_ISSUE")
	jiraHost := os.Getenv("JIRA_HOST")

	if authToken == "" {
		fmt.Println("Need to set JIRA_AUTH_TOKEN in .env")
		return
	}

	jiraClient := jira.NewClient(jiraHost, authToken)
	if err := jiraClient.CheckAuth(); err != nil {
		fmt.Printf("Check jira auth error: %s", err.Error())
		return
	}

	command := "worklog"

	argsWithProg := os.Args
	interactive := false
	for i, arg := range argsWithProg {
		if arg == "-i" {
			interactive = true
			continue
		}
		if arg == "-command" && i != len(argsWithProg)-1 {
			command = argsWithProg[i+1]
		}
	}

	if command == "worklog" {
		for {
			addWorklog(interactive, meetingsJiraIssue, jiraClient)
		}
	}
	if command == "my-issues" {
		myIssues(jiraClient)
		return
	}
	fmt.Printf("Unknown command: %s", command)
}

func addWorklog(interactive bool, meetingsJiraIssue string, jiraClient *jira.Client) {
	defer fmt.Printf("\n\n")

	var (
		issue, date, comment string
		spentTime            time.Duration
		err                  error
	)

	if interactive {
		issue, date, spentTime, comment, err = prepareArgsInteractive(meetingsJiraIssue)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	} else {
		issue, date, spentTime, comment = prepareArgs()
	}

	send := checkSend(issue, date, spentTime, comment)
	if !send {
		fmt.Printf("Result wasn't sent")
		return
	}

	err = jiraClient.AddWorklog(issue, date, spentTime, comment)
	if err != nil {
		fmt.Printf("Result wasn't sent: %s", err.Error())
	}
	return
}

func myIssues(jiraClient *jira.Client) {
	issues, err := jiraClient.EverAssignedIssues()
	if err != nil {
		fmt.Printf("Can't get issues: %s", err.Error())
		return
	}
	formattedIssues := lo.Map(issues, func(issue jira.Issue, _ int) string {
		return fmt.Sprintf("%s: %s", issue.Key, issue.Fields.Summary)
	})
	fmt.Printf(strings.Join(formattedIssues, "\n"))
}

func prepareArgsInteractive(meetingsJiraIssue string) (string, string, time.Duration, string, error) {
	fmt.Print("--- Interactive mode ---\n")
	reader := bufio.NewReader(os.Stdin)

	issue, err := read(reader, "Issue: ")
	if err != nil {
		panic(err)
	}
	if issue == "" {
		if meetingsJiraIssue == "" {
			return "", "", time.Second, "", errors.New("Please input jira issue or set MEETINGS_JIRA_ISSUE in .env")
		}
		issue = meetingsJiraIssue
	}
	date, err := read(reader, "Date: ")
	if err != nil {
		return "", "", time.Second, "", err
	}
	if date == "" {
		date = time.Now().Format(time.DateOnly)
	}
	var spentTime time.Duration
	spent, err := read(reader, "Spent time: ")
	if err != nil {
		return "", "", time.Second, "", err
	}
	if spentHours, err := strconv.Atoi(spent); err == nil {
		spentTime = time.Duration(spentHours) * time.Hour
	} else {
		spentTime, err = time.ParseDuration(spent)
		if err != nil {
			return "", "", time.Second, "", err
		}
	}
	comment, err := read(reader, "Comment: ")
	if err != nil {
		return "", "", time.Second, "", err
	}
	return issue, date, spentTime, comment, nil
}

func read(reader *bufio.Reader, text string) (string, error) {
	fmt.Print(text)
	result, _ := reader.ReadString('\n')
	result = strings.TrimRight(result, "\r\n")
	return result, nil
}

func checkSend(issue string, date string, spentTime time.Duration, comment string) bool {
	fmt.Println(issue, date, spentTime, comment)

	reader := bufio.NewReader(os.Stdin)
	answer, err := read(reader, "Send Y/n: ")
	if err != nil || answer == "n" || answer == "N" {
		return false
	}
	return true
}

func prepareArgs() (string, string, time.Duration, string) {
	fmt.Print("--- Bulk mode ---\n")
	fmt.Print("Format {YYYY-MM-DD} {ISSUE} {SPENT_TIME} {COMMENT}\n")
	reader := bufio.NewReader(os.Stdin)

	row, err := read(reader, "Worklog row: ")
	if err != nil {
		panic(err)
	}
	splitted := strings.Split(row, " ")
	if len(splitted) < 3 {
		panic("wrong worklog!")
	}

	date := splitted[0]
	issue := splitted[1]
	spent := splitted[2]
	spentTime, err := time.ParseDuration(spent)
	if err != nil {
		panic(err)
	}
	comment := ""
	if len(splitted) > 3 {
		comment = strings.Join(splitted[3:], " ")
	}

	return issue, date, spentTime, comment
}
