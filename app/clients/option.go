package clients

import (
	"github.com/temporalio/temporal-jumpstart-golang/app/clients/temporal"
)

type Option func(*Clients)

func WithTemporalOptions(namespace string, opts ...temporal.Option) Option {
	return func(clients *Clients) {
		clients.temporalOptions[namespace] = append(clients.temporalOptions[namespace], opts...)
	}
}
