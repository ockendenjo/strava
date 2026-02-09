package strava

import "fmt"

type HttpStatusError struct {
	StatusCode int
	Body       string
}

func (e HttpStatusError) Error() string {
	return fmt.Sprintf("HTTP request returned status code %d: %s", e.StatusCode, e.Body)
}
