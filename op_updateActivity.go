package strava

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *client) UpdateActivity(ctx context.Context, id int64, updates ActivityUpdates) error {
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	jsonBody, err := json.Marshal(updates)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", "https://www.strava.com/api/v3/activities/"+fmt.Sprint(id), bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return HttpStatusError{StatusCode: res.StatusCode}
	}

	return nil
}