package api

import (
	"context"
	"errors"
	"github.com/temporalio/temporal-jumpstart-golang/app/api/encoding"
	"github.com/temporalio/temporal-jumpstart-golang/app/api/messages"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/workflows"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"log"
	"net/http"

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
		var output *string
		if err := result.Get(&output); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte(*output))
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
			ID:                       workflowId,
			TaskQueue:                deps.Config.Temporal.Worker.TaskQueue,
			WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
			WorkflowIDConflictPolicy: enums.WORKFLOW_ID_CONFLICT_POLICY_FAIL,
		}

		_, err := deps.Clients.Temporal.ExecuteWorkflow(r.Context(), options, workflows.Ping, body.Ping)
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

		w.Header().Set("Location", "./"+workflowId)
		w.WriteHeader(http.StatusAccepted)
	}).Methods(http.MethodPut)

	return router
}
