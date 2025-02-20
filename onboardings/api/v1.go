package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/api/encoding"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/api/messages"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/clients"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows/onboardings"
	apiv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/api/v1"
	queriesv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/queries/v1"
	workflowsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/workflows/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"go.temporal.io/sdk/client"
)

type V1Dependencies struct {
	Clients *clients.Clients
	Config  *config.Config
}

func createV1Router(ctx context.Context, deps *V1Dependencies, router *mux.Router) *mux.Router {

	router.HandleFunc("/pings/{id}", func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		workflowId := vars["id"]

		result, err := deps.Clients.Temporal.QueryWorkflow(
			r.Context(),
			workflowId,
			"",
			workflows.QueryPing)
		if err != nil {
			if _, ok := err.(*serviceerror.NotFound); ok {
				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		var output string
		if err := result.Get(&output); err != nil {
			log.Printf("Error getting workflow output: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte(output))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

	}).Methods(http.MethodGet)

	router.HandleFunc("/pings/{id}", func(w http.ResponseWriter, r *http.Request) {
		var body *messages.PutPing
		if err := encoding.DecodeJSONBody(w, r, &body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		workflowId := vars["id"]

		options := client.StartWorkflowOptions{
			ID:        workflowId,
			TaskQueue: deps.Config.Temporal.Worker.TaskQueue,
			// This configures how to deal with prior attempts of an Onboarding.
			// We want to allow a "do over" of the Onboarding only if the prior attempts failed.
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
			// This configures how to deal with any Running Onboarding currently in progress.
			// We want to fail if already in flight.
			WorkflowIDConflictPolicy: enums.WORKFLOW_ID_CONFLICT_POLICY_FAIL,
			// This makes explicit when our Onboarding could not be started due to the configuration above.
			// Otherwise, this could silently fail and it becomes harder to track down.
			WorkflowExecutionErrorWhenAlreadyStarted: true,
		}

		wfRun, err := deps.Clients.Temporal.ExecuteWorkflow(r.Context(), options, workflows.Ping, body.Ping)
		if err != nil {
			var alreadyStartedErr *serviceerror.WorkflowExecutionAlreadyStarted
			if errors.As(err, &alreadyStartedErr) {
				http.Error(w, "Workflow already exists", http.StatusConflict)
				return
			}

			log.Printf("Failed to execute workflow '%s': %v", wfRun.GetID(), err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		host := r.Header.Get("Host")
		if host == "" {
			host = r.Host
		}
		host = fmt.Sprintf("https://%s", host)
		fmt.Println(host)
		link, err := url.Parse(host)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		location := link.ResolveReference(r.URL)

		w.Header().Set("Location", location.String())
		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodPut)

	router.HandleFunc("/onboardings/{id}", func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		workflowId := vars["id"]

		result, err := deps.Clients.Temporal.QueryWorkflow(
			r.Context(),
			workflowId,
			"",
			onboardings.QueryEntityOnboardingState)
		if err != nil {
			if _, ok := err.(*serviceerror.NotFound); ok {
				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		var queryResult *queriesv1.EntityOnboardingStateResponse
		if err := result.Get(&queryResult); err != nil {
			log.Printf("Error getting workflow queryResult: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		var output = &apiv1.OnboardingsGet{
			Id:           queryResult.Id,
			CurrentValue: queryResult.SentRequest.Value,
			Approval:     queryResult.Approval,
		}
		err = encoding.EncodeJSONResponseBody(w, output, http.StatusOK)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

	}).Methods(http.MethodGet)

	router.HandleFunc("/onboardings/{id}", func(w http.ResponseWriter, r *http.Request) {
		var body *apiv1.OnboardingsPut
		if err := encoding.DecodeJSONBody(w, r, &body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		workflowId := vars["id"]

		options := client.StartWorkflowOptions{
			ID:                                       workflowId,
			TaskQueue:                                deps.Config.Temporal.Worker.TaskQueue,
			WorkflowIDReusePolicy:                    enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
			WorkflowIDConflictPolicy:                 enums.WORKFLOW_ID_CONFLICT_POLICY_FAIL,
			WorkflowExecutionErrorWhenAlreadyStarted: true,
		}

		params := &workflowsv1.OnboardEntityRequest{
			Id:                       workflowId,
			Value:                    body.Value,
			CompletionTimeoutSeconds: 0,
			DeputyOwnerEmail:         "",
			SkipApproval:             false,
		}

		_, err := deps.Clients.Temporal.ExecuteWorkflow(r.Context(), options, onboardings.TypeWorkflowOnboardEntity, params)
		if err != nil {
			var alreadyStartedErr *serviceerror.WorkflowExecutionAlreadyStarted
			if errors.As(err, &alreadyStartedErr) {
				http.Error(w, "Workflow already exists", http.StatusConflict)
				return
			}

			log.Printf("Failed to execute workflow '%s': %v", workflowId, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		host := r.Header.Get("Host")
		if host == "" {
			host = r.Host
		}
		host = fmt.Sprintf("https://%s", host)
		fmt.Println(host)
		link, err := url.Parse(host)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		location := link.ResolveReference(r.URL)

		w.Header().Set("Location", location.String())
		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodPut)

	return router
}
