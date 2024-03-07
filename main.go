package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"jira-helper/jira_client"
)

func main() {
	authToken := os.Getenv("JIRA_AUTH_TOKEN")
	meetingsJiraIssue := os.Getenv("MEETINGS_JIRA_ISSUE")
	jiraHost := os.Getenv("JIRA_HOST")

	if authToken == "" {
		fmt.Println("Need to set JIRA_AUTH_TOKEN in .env")
		return
	}

	jiraClient := jira_client.NewClient(jiraHost, authToken)

	if err := jiraClient.CheckAuth(); err != nil {
		fmt.Printf("Check jira auth error: %s", err.Error())
		return
	}

	argsWithProg := os.Args
	interactive := false
	for _, arg := range argsWithProg {
		if arg == "-i" {
			interactive = true
		}
	}

	for {
		addWorklog(interactive, meetingsJiraIssue, jiraClient)
	}
}

func addWorklog(interactive bool, meetingsJiraIssue string, jiraClient *jira_client.Client) {
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
