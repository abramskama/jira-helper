package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	JIRA_TOKEN          = "персональный токен из жиры"
	MEETINGS_JIRA_ISSUE = "INT-18"
)

func main() {
	argsWithProg := os.Args
	interactive := false
	for _, arg := range argsWithProg {
		if arg == "-i" {
			interactive = true
		}
	}

	if interactive {
		issue, date, spentTime, comment := prepareArgsInteractive()
		logWork(issue, date, spentTime, comment)
		return
	}

	for {
		issue, date, spentTime, comment := prepareArgs()
		logWork(issue, date, spentTime, comment)
	}
}

func logWork(issue string, date string, spentTime time.Duration, comment string) {
	req, err := addWorklogRequest(issue, date, spentTime, comment)
	if err != nil {
		panic(err)
	}

	send := checkSend(issue, date, spentTime, comment)
	if !send {
		return
	}

	err = doRequest(req)
	if err != nil {
		panic(err)
	}
}

func prepareArgsInteractive() (string, string, time.Duration, string) {
	fmt.Print("Выбран запуск в интерактивном режиме")
	reader := bufio.NewReader(os.Stdin)

	issue, err := read(reader, "Issue: ")
	if err != nil {
		panic(err)
	}
	if issue == "" {
		issue = MEETINGS_JIRA_ISSUE
	}
	date, err := read(reader, "Date: ")
	if err != nil {
		panic(err)
	}
	if date == "" {
		date = time.Now().Format(time.DateOnly)
	}
	var spentTime time.Duration
	spent, err := read(reader, "Spent time: ")
	if err != nil {
		panic(err)
	}
	if spentHours, err := strconv.Atoi(spent); err == nil {
		spentTime = time.Duration(spentHours) * time.Hour
	} else {
		spentTime, err = time.ParseDuration(spent)
		if err != nil {
			panic(err)
		}
	}
	comment, err := read(reader, "Comment: ")
	if err != nil {
		panic(err)
	}
	return issue, date, spentTime, comment
}

func read(reader *bufio.Reader, text string) (string, error) {
	fmt.Print(text)
	result, _ := reader.ReadString('\n')
	result = strings.TrimRight(result, "\r\n")
	return result, nil
}

func doRequest(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		if resp.Body != nil {
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			var response json.RawMessage
			_ = json.Unmarshal(respBody, &response)
			log.Println(string(respBody))
		}
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	_, err = ioutil.ReadAll(resp.Body)
	return err
}

func addWorklogRequest(issue string, date string, spentTime time.Duration, comment string) (*http.Request, error) {
	body := `{"timeSpentSeconds": "%d",	"started": "%sT18:00:00.751+0000", "comment": "%s"}`
	body = fmt.Sprintf(body, int(spentTime.Seconds()), date, comment)
	fmt.Println(body)
	buf := strings.NewReader(body)
	url := fmt.Sprintf("https://jira.lamoda.ru/rest/api/2/issue/%s/worklog", issue)

	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+JIRA_TOKEN)
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func checkSend(issue string, date string, spentTime time.Duration, comment string) bool {
	fmt.Println(issue, date, spentTime, comment)

	reader := bufio.NewReader(os.Stdin)
	answer, err := read(reader, "Send y/N: ")
	if err != nil || answer == "N" {
		return false
	}
	return true
}

func prepareArgs() (string, string, time.Duration, string) {
	fmt.Print("Выбран запуск в построчном режиме, для интерактивного используйте флаг -i")
	fmt.Print("Формат {YYYY-MM-DD} {ISSUE} {SPENT_TIME} {COMMENT}")
	reader := bufio.NewReader(os.Stdin)

	row, err := read(reader, "Raw worklog: ")
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
