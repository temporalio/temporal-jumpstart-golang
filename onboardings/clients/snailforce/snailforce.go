package snailforce

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/snailforce/v1/snailforcev1connect"
	"net/http"
	"net/url"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewClient(_ context.Context, cfg *config.Config, httpClient Doer) (snailforcev1connect.SnailforceServiceClient, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	parsedUrl, err := url.Parse(cfg.Snailforce.URL)
	if err != nil {
		return nil, err
	}
	snailforceClient := snailforcev1connect.NewSnailforceServiceClient(httpClient, parsedUrl.String())
	return snailforceClient, nil
}
