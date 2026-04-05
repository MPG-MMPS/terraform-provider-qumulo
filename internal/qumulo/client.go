package qumulo

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Method int

const (
	GET Method = iota + 1
	PUT
	POST
	PATCH
	DELETE
)

func (m Method) String() string {
	return [...]string{"GET", "PUT", "POST", "PATCH", "DELETE"}[m-1]
}

type Client struct {
	HostURL     string
	HTTPClient  *http.Client
	BearerToken string
	Auth        AuthStruct
}

type AuthStruct struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	BearerToken string `json:"bearer_token"`
}

func NewClient(ctx context.Context, host, port, username, password *string) (*Client, error) {
	if host == nil || port == nil || username == nil || password == nil {
		return nil, fmt.Errorf("cannot create client: host, port, username, and password must be set")
	}

	hostURL := fmt.Sprintf("https://%s:%s", *host, *port)

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second, Transport: transCfg},
		HostURL:    hostURL,
		Auth: AuthStruct{
			Username: *username,
			Password: *password,
		},
	}

	ar, err := c.SignIn(ctx)
	if err != nil {
		return nil, err
	}

	c.BearerToken = ar.BearerToken
	c.HostURL = hostURL

	tflog.Info(ctx, "Qumulo client configured", map[string]interface{}{
		"host":     host,
		"port":     port,
		"username": username,
	})

	return &c, nil
}

func DoRequest[RQ interface{}, R interface{}](ctx context.Context, client *Client, method Method, endpointURI string, reqBody *RQ) (*R, error) {
	bearerToken := "Bearer " + client.BearerToken
	hostURL := client.HostURL

	var parsedReqBody io.Reader
	if reqBody != nil {
		rb, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		parsedReqBody = strings.NewReader(string(rb))
	}

	url := fmt.Sprintf("%s%s", hostURL, endpointURI)
	req, err := http.NewRequestWithContext(ctx, method.String(), url, parsedReqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", bearerToken)
	req.Header.Add("Content-Type", "application/json")

	tflog.Trace(ctx, "Executing API request", map[string]interface{}{
		"url":    url,
		"method": method.String(),
	})

	body, err := client.makeHTTPRequest(req)
	if err != nil {
		return nil, err
	}

	var cr R
	if len(body) == 0 {
		return nil, nil
	}

	err = json.Unmarshal(body, &cr)
	if err != nil {
		return nil, err
	}

	return &cr, nil
}

func (c *Client) makeHTTPRequest(req *http.Request) ([]byte, error) {
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
