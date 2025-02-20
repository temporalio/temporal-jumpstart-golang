package snailforce

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/snailforce/v1/snailforcev1connect"
	"net/http"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewClient(ctx context.Context, cfg *config.Config, httpClient Doer) (snailforcev1connect.SnailforceServiceClient, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	snailforceClient := snailforcev1connect.NewSnailforceServiceClient(httpClient, cfg.Snailforce.URL.String())
	return snailforceClient, nil
}
