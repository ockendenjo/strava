package strava

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
)

func (c *client) Authorize(ctx context.Context, code string) error {
	gotParams, err := c.psClient.GetParams(ctx)
	if err != nil {
		return err
	}
	c.params = &gotParams

	u, err := url.Parse("https://www.strava.com/oauth/token")
	if err != nil {
		log.Fatal(err)
	}

	q := u.Query()
	q.Set("client_id", c.params.ClientId)
	q.Set("client_secret", c.params.ClientSecret)
	q.Set("grant_type", "authorization_code")
	q.Set("code", code)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), nil)
	if err != nil {
		return err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return HttpStatusError{StatusCode: res.StatusCode, Body: string(bytes)}
	}

	var refreshRes refreshResponse
	err = json.Unmarshal(bytes, &refreshRes)
	if err != nil {
		return err
	}

	c.params.AccessToken = refreshRes.AccessToken
	c.params.RefreshToken = refreshRes.RefreshToken
	c.params.ExpiryTime = refreshRes.ExpiryTime

	//Set SSM params
	err = c.psClient.SetRefreshedParams(ctx, refreshRes.RefreshToken, refreshRes.AccessToken, refreshRes.ExpiryTime)
	return err
}
