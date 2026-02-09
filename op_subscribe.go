package strava

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func (c *client) Subscribe(ctx context.Context, callbackURL, verifyToken string) error {
	params, err := c.psClient.GetParams(ctx)
	if err != nil {
		return err
	}

	err = c.psClient.SetVerifyToken(ctx, verifyToken)
	if err != nil {
		return err
	}

	u, err := url.Parse("https://www.strava.com/api/v3/push_subscriptions")
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("client_id", params.ClientId)
	q.Set("client_secret", params.ClientSecret)
	q.Set("callback_url", callbackURL)
	q.Set("verify_token", verifyToken)
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

	if res.StatusCode == http.StatusCreated {
		return nil
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return HttpStatusError{StatusCode: res.StatusCode, Body: string(b)}
}
