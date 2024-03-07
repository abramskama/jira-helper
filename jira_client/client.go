package jira_client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	host      string
	authToken string
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
	uri := fmt.Sprintf("/rest/api/2/issue/%s/worklog", issue)

	body := `{"timeSpentSeconds": "%d",	"started": "%sT18:00:00.751+0000", "comment": "%s"}`
	body = fmt.Sprintf(body, int(spentTime.Seconds()), date, comment)

	return c.Do(http.MethodPost, uri, body)
}

func (c *Client) Do(method, uri, body string) error {
	req, err := c.request(method, uri, body)
	if err != nil {
		return err
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

func doRequest(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > http.StatusIMUsed {
		body := ""
		if resp.Body != nil {
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("fail to read response body: %s", err.Error())
			}
			body = string(respBody)
		}
		return fmt.Errorf("status code %d, body: %s", resp.StatusCode, body)
	}
	_, err = ioutil.ReadAll(resp.Body)
	return err
}
