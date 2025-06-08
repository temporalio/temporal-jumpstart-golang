package main

import (
	"context"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/api"
	appclients "github.com/temporalio/temporal-jumpstart-golang/onboardings/clients"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	failure2 "go.temporal.io/api/failure/v1"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/proto"
	"log"
	"net/http"
)

func main() {
	createFailure := func() *failure2.Failure {
		return &failure2.Failure{
			Message:           "a",
			Source:            "a",
			StackTrace:        "a",
			EncodedAttributes: nil,
			Cause:             nil,
			FailureInfo: &failure2.Failure_ApplicationFailureInfo{ApplicationFailureInfo: &failure2.ApplicationFailureInfo{
				Type: "application_failure",
			}},
		}
	}
	failure := createFailure()
	str := "ABCDEFGHIKLMNOPQRSTUVWX"
	f2 := &workflowservice.RespondWorkflowTaskFailedRequest{
		TaskToken:      nil,
		Cause:          0,
		Failure:        failure,
		Identity:       str,
		BinaryChecksum: str,
		Namespace:      str,
		Messages:       nil,
		WorkerVersion:  nil,
		Deployment:     nil,
	}
	cur := f2.Failure
	for i := 0; i < 100; i++ {
		cur.Cause = createFailure()
		cur = cur.Cause
	}
	bytes, err := proto.Marshal(f2)
	fmt.Println("byte count", len(bytes))
	ctx := context.Background()
	cfg := config.MustNewConfig()
	clients, err := appclients.NewClients(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	apiRouter, err := api.CreateAPIRouter(ctx, cfg, clients)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(cfg.API.URL.Host, apiRouter))
}
