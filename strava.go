package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/ockendenjo/strava/services/ps"
)

type Client interface {
	Authorize(ctx context.Context, code string) error
	GetActivities(ctx context.Context, page int) ([]Activity, error)
	GetActivity(ctx context.Context, id int64) (*Activity, error)
	GetActivityStream(ctx context.Context, id int64) (*ActivityStream, error)
	Subscribe(ctx context.Context, callbackURL, verifyToken string) error
}

func NewClient(ssmClient *ssm.Client, httpClient *http.Client) Client {
	psClient := ps.NewParamsClient(ssmClient)
	return &client{psClient: psClient, httpClient: httpClient}
}

type client struct {
	psClient   ps.ParamsClient
	httpClient *http.Client
	params     *ps.StravaParams
}

func (c *client) getAccessToken(ctx context.Context) (string, error) {
	if c.params == nil {
		gotParams, err := c.psClient.GetParams(ctx)
		if err != nil {
			return "", err
		}
		c.params = &gotParams
	}

	if c.params.ExpiryTime > time.Now().Unix() {
		return c.params.AccessToken, nil
	}

	//Need to refresh
	u, err := url.Parse("https://www.strava.com/oauth/token")
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("client_id", c.params.ClientId)
	q.Set("client_secret", c.params.ClientSecret)
	q.Set("grant_type", "refresh_token")
	q.Set("refresh_token", c.params.RefreshToken)
	u.RawQuery = q.Encode()
	fmt.Println(u)

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), nil)
	if err != nil {
		return "", err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", HttpStatusError{StatusCode: res.StatusCode, Body: string(bytes)}
	}

	var refreshRes refreshResponse
	err = json.Unmarshal(bytes, &refreshRes)
	if err != nil {
		return "", err
	}

	c.params.AccessToken = refreshRes.AccessToken
	c.params.RefreshToken = refreshRes.RefreshToken
	c.params.ExpiryTime = refreshRes.ExpiryTime

	//Set SSM params
	err = c.psClient.SetRefreshedParams(ctx, refreshRes.RefreshToken, refreshRes.AccessToken, refreshRes.ExpiryTime)
	if err != nil {
		return "", err
	}
	return refreshRes.AccessToken, nil
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiryTime   int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}
