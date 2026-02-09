package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *client) GetActivityStream(ctx context.Context, id int64) (*ActivityStream, error) {
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://www.strava.com/api/v3/activities/%d/streams?keys=latlng&key_by_type=true", id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	var actStream ActivityStream
	err = json.Unmarshal(bytes, &actStream)
	return &actStream, err
}
