package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
)

func (c *QuayClient) GetRobotAccount(accountType string, accountName string, robotName string) (RobotAccount, *http.Response, QuayApiError) {

	req, err := c.newRequest("GET", fmt.Sprintf("/api/v1/%s/%s/robots/%s", accountType, accountName, robotName), nil)
	if err != nil {
		return RobotAccount{}, nil, QuayApiError{Error: err}
	}
	var getRobotResponse RobotAccount
	resp, err := c.do(req, &getRobotResponse)

	return getRobotResponse, resp, QuayApiError{Error: err}
}

func (c *QuayClient) CreateRobotAccount(accountType string, accountName string, robotName string) (RobotAccount, *http.Response, QuayApiError) {

	req, err := c.newRequest("PUT", fmt.Sprintf("/api/v1/%s/%s/robots/%s", accountType, accountName, robotName), nil)
	if err != nil {
		return RobotAccount{}, nil, QuayApiError{Error: err}
	}
	var createRobotResponse RobotAccount
	resp, err := c.do(req, &createRobotResponse)

	return createRobotResponse, resp, QuayApiError{Error: err}
}

func (c *QuayClient) DeleteRobotAccount(accountType string, accountName string, robotName string) (*http.Response, QuayApiError) {

	req, err := c.newRequest("DELETE", fmt.Sprintf("/api/v1/%s/%s/robots/%s", accountType, accountName, robotName), nil)
	if err != nil {
		return nil, QuayApiError{Error: err}
	}
	resp, err := c.do(req, nil)

	return resp, QuayApiError{Error: err}
}

func (c *QuayClient) CreateTeam(accountName string, team *Team) (Team, *http.Response, QuayApiError) {

	req, err := c.newRequest("PUT", fmt.Sprintf("/api/v1/organization/%s/team/%s", accountName, team.Name), team)
	if err != nil {
		return Team{}, nil, QuayApiError{Error: err}
	}
	var createTeamResponse Team
	resp, err := c.do(req, &createTeamResponse)

	return createTeamResponse, resp, QuayApiError{Error: err}
}

func (c *QuayClient) AddTeamMember(accountName, teamName, memberName string) (*http.Response, QuayApiError) {

	req, err := c.newRequest("PUT", fmt.Sprintf("/api/v1/organization/%s/team/%s/members/%s", accountName, teamName, memberName), nil)
	if err != nil {
		return nil, QuayApiError{Error: err}
	}
	resp, err := c.do(req, nil)

	return resp, QuayApiError{Error: err}
}

func (c *QuayClient) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.baseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)

	if !isZeroOfUnderlyingType(c.authToken) {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}
func (c *QuayClient) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if v != nil {

		if _, ok := v.(*StringValue); ok {
			responseData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return resp, err
			}
			responseObject := v.(*StringValue)
			responseObject.Value = string(responseData)

		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return resp, err
			}
		}

	}

	return resp, err
}

func NewClient(httpClient *http.Client, baseUrl string, authToken string) (*QuayClient, error) {
	quayClient := QuayClient{
		httpClient: httpClient,
		authToken:  authToken,
	}

	parsedUrl, err := url.Parse(baseUrl)

	if err != nil {
		return nil, err
	}

	quayClient.baseURL = parsedUrl

	return &quayClient, nil
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}