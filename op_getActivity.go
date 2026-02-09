package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *client) GetActivity(ctx context.Context, id int64) (*Activity, error) {
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.strava.com/api/v3/activities/"+fmt.Sprint(id), nil)
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

	if res.StatusCode != http.StatusOK {
		return nil, HttpStatusError{StatusCode: res.StatusCode}
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var activity Activity
	err = json.Unmarshal(bytes, &activity)
	return &activity, err
}
