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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, stravaUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

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
