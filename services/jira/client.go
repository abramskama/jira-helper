package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	host      string
	authToken string
}

type IssuesResponse struct {
	Issues []Issue `json:"issues"`
}

type Issue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
	} `json:"fields"`
}

func NewClient(host, authToken string) *Client {
	return &Client{
		host:      host,
		authToken: authToken,
	}
}

func (c *Client) CheckAuth() error {
	return c.Do(http.MethodGet, "/rest/api/2/myself", "")
}

func (c *Client) AddWorklog(issue string, date string, spentTime time.Duration, comment string) error {
	date, err := convertDate(date)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("/rest/api/2/issue/%s/worklog", issue)

	body := `{"timeSpentSeconds": "%d",	"started": "%sT18:00:00.751+0000", "comment": "%s"}`
	body = fmt.Sprintf(body, int(spentTime.Seconds()), date, comment)

	return c.Do(http.MethodPost, uri, body)
}

func (c *Client) EverAssignedIssues() ([]Issue, error) {
	body, err := c.DoResponse(http.MethodGet, "/rest/api/2/search?jql=assignee%20was%20in%20(currentUser())%20ORDER%20BY%20updated%20DESC", "")
	if err != nil {
		return nil, err
	}

	issuesResponse := IssuesResponse{}
	err = json.Unmarshal(body, &issuesResponse)
	if err != nil {
		return nil, err
	}

	//if len(issuesResponse.Issues) > 0 {
	//	_, _ = c.TIS(issuesResponse.Issues[0].Key)
	//}

	issues := make([]string, 0, len(issuesResponse.Issues))
	for _, issue := range issuesResponse.Issues {
		issues = append(issues, issue.Key)
	}
	return issuesResponse.Issues, nil
}

func (c *Client) Issue(issue string) (json.RawMessage, error) {
	uri := fmt.Sprintf("/rest/api/2/issue/%s", issue)
	body, err := c.DoResponse(http.MethodGet, uri, "")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println(string(body))

	resp := json.RawMessage{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	log.Println(resp)
	return resp, nil
}

func (c *Client) TIS(issue string) (json.RawMessage, error) {
	log.Println("test")
	uri := fmt.Sprintf("/rest/tis/report/1.0/api/issue?issueKey=%s&columnsBy=%s", issue, "firsttransitionfromstatusdate")
	body, err := c.DoResponse(http.MethodGet, uri, "")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println(string(body))

	resp := json.RawMessage{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	log.Println(resp)
	return resp, nil
}

func (c *Client) Do(method, uri, body string) error {
	_, err := c.DoResponse(method, uri, body)
	return err
}

func (c *Client) DoResponse(method, uri, body string) ([]byte, error) {
	req, err := c.request(method, uri, body)
	if err != nil {
		return nil, err
	}

	return doRequest(req)
}

func (c *Client) request(method, uri, body string) (*http.Request, error) {
	url := fmt.Sprintf("https://%s%s", c.host, uri)

	buf := strings.NewReader(body)
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.authToken)
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func doRequest(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > http.StatusIMUsed {
		body := ""
		if resp.Body != nil {
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("fail to read response body: %s", err.Error())
			}
			body = string(respBody)
		}
		return nil, fmt.Errorf("status code %d, body: %s", resp.StatusCode, body)
	}
	return io.ReadAll(resp.Body)
}
