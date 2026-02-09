package ps

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

const ssmPrefix = "/strava/"

func NewParamClient(client *ssm.Client) *ParamClient {
	return &ParamClient{ssmClient: client}
}

type ParamClient struct {
	ssmClient *ssm.Client
	params    map[string]string
}

func (c *ParamClient) GetParams(ctx context.Context) (map[string]string, error) {
	if c.params != nil {
		return c.params, nil
	}

	result, err := c.ssmClient.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
		Path:           aws.String(ssmPrefix),
		WithDecryption: aws.Bool(true),
		Recursive:      aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	paramMap := map[string]string{}

	for _, parameter := range result.Parameters {
		name := strings.Replace(*parameter.Name, ssmPrefix, "", 1)
		paramMap[name] = *parameter.Value
	}

	c.params = paramMap
	return paramMap, nil
}

func (c *ParamClient) SetParam(ctx context.Context, key string, value string, paramType types.ParameterType) error {
	_, err := c.ssmClient.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String(ssmPrefix + key),
		Value:     aws.String(value),
		Type:      paramType,
		Overwrite: aws.Bool(true),
	})
	return err
}
