package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *client) GetActivities(ctx context.Context, page int) ([]Activity, error) {
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	stravaUrl := fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?page=%d", page)
	req, err := http.NewRequestWithContext(ctx, "GET", stravaUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, HttpStatusError{StatusCode: res.StatusCode, Body: string(bytes)}
	}

	var activities []Activity
	err = json.Unmarshal(bytes, &activities)
	return activities, err
}
