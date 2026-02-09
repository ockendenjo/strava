package strava

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	ssmTypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/ockendenjo/handler"
	"github.com/ockendenjo/strava/services/ps"
)

func NewClient(paramClient *ps.ParamClient, httpClient *http.Client) *StravaClient {
	return &StravaClient{paramClient: paramClient, httpClient: httpClient}
}

func (sc *StravaClient) GetAccessToken(ctx *handler.Context) (string, error) {
	if sc.accessToken != nil && sc.expiryTime < time.Now().Unix() {
		return *sc.accessToken, nil
	}

	logger := ctx.GetLogger()
	params, err := sc.paramClient.GetParams(ctx)
	if err != nil {
		return "", err
	}

	expiryTime, err := strconv.ParseInt(params["expiryTime"], 10, 64)
	if err != nil {
		expiryTime = 0
	}
	nowSeconds := time.Now().Unix()
	if expiryTime > nowSeconds {
		return params["accessToken"], nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.strava.com/oauth/token", nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("client_id", params["clientId"])
	q.Add("client_secret", params["clientSecret"])
	q.Add("grant_type", "refresh_token")
	q.Add("refresh_token", params["refreshToken"])
	req.URL.RawQuery = q.Encode()

	response, err := sc.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		logger.Error("Refreshing access token returned error", "body", string(bytes), "statusCode", response.StatusCode)
		return "", errors.New("non-200 when refreshing access token")
	}

	var authRes authResponse
	err = json.Unmarshal(bytes, &authRes)
	if err != nil {
		return "", err
	}

	c := make(chan channelRes, 3)
	setParam := func(ctx context.Context, key string, value string, paramType ssmTypes.ParameterType, c chan channelRes) {
		err = sc.paramClient.SetParam(ctx, key, value, paramType)
		if err != nil {
			c <- channelRes{err: err}
		} else {
			c <- channelRes{err: nil}
		}
	}
	go setParam(ctx, "refreshToken", authRes.RefreshToken, ssmTypes.ParameterTypeSecureString, c)
	go setParam(ctx, "accessToken", authRes.AccessToken, ssmTypes.ParameterTypeSecureString, c)
	go setParam(ctx, "expiryTime", fmt.Sprintf("%d", authRes.ExpiresAt), ssmTypes.ParameterTypeString, c)

	var chanErr error
	for i := 0; i < 3; i++ {
		res := <-c
		if res.err != nil {
			chanErr = res.err
		}
	}

	sc.accessToken = &authRes.AccessToken
	sc.expiryTime = authRes.ExpiresAt

	return authRes.AccessToken, chanErr
}

func (sc *StravaClient) GetActivityStream(ctx *handler.Context, id int64) (*ActivityStream, error) {
	accessToken, err := sc.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://www.strava.com/api/v3/activities/%d/streams?keys=latlng&key_by_type=true", id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := sc.httpClient.Do(req)
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
		return nil, HttpStatusError{StatusCode: res.StatusCode, Body: bytes}
	}

	var actStream ActivityStream
	err = json.Unmarshal(bytes, &actStream)
	return &actStream, err
}

func (sc *StravaClient) GetActivity(ctx *handler.Context, id int64) (*Activity, error) {
	accessToken, err := sc.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.strava.com/api/v3/activities/"+fmt.Sprint(id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := sc.httpClient.Do(req)
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

type channelRes struct {
	err error
}

type StravaClient struct {
	paramClient *ps.ParamClient
	httpClient  *http.Client
	accessToken *string
	expiryTime  int64
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

type Activity struct {
	ID        int64       `json:"id"`
	Name      string      `json:"name"`
	Map       PolylineMap `json:"map"`
	SportType string      `json:"sport_type"`
	GearID    string      `json:"gear_id"`
	StartDate string      `json:"start_date"`
}

type ActivityStream struct {
	LatLng LatLngData `json:"latlng"`
}

type LatLngData struct {
	Data [][]float64 `json:"data"`
}

type PolylineMap struct {
	ID       string `json:"id"`
	Polyline string `json:"polyline"`
}

type HttpStatusError struct {
	StatusCode int
	Body       []byte
}

func (e HttpStatusError) Error() string {
	return fmt.Sprintf("HTTP request returned status code %d: %s", e.StatusCode, string(e.Body))
}
