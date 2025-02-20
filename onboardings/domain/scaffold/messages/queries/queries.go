package queries

import (
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/scaffold/messages/values"
)

type QueryActivityRequest struct {
	ID string `json:"id"`
}

type QueryActivityResponse struct {
	ID        string           `json:"id"`
	Value     string           `json:"value"`
	ValidFrom values.DateRange `json:"validFrom"`
}
