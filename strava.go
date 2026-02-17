package strava

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/ockendenjo/strava/services/ps"
)

type Client interface {
	Authorize(ctx context.Context, code string) error
	GetAccessToken(ctx context.Context) (string, error)
	GetActivities(ctx context.Context, page int) ([]Activity, error)
	GetActivity(ctx context.Context, id int64) (*Activity, error)
	GetActivityStream(ctx context.Context, id int64) (*ActivityStream, error)
	Subscribe(ctx context.Context, callbackURL, verifyToken string) error
	UpdateActivity(ctx context.Context, id int64, updates ActivityUpdates) error
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
