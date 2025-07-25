package clients

import "github.com/temporalio/temporal-jumpstart-golang/app/clients/temporal"

type Option func(*Clients)

func WithTemporalOptions(opts ...[]temporal.Option) Option {
	return func(clients *Clients) {
		clients.temporalOptions = opts
	}
}
func WithTemporalClientCount(count int) Option {
	return func(clients *Clients) {
		clients.temporalClientCount = count
	}
}
