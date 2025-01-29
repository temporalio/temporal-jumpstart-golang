package api

import (
	"encoding/json"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"go.temporal.io/api/serviceerror"
	"net/http"

	"github.com/gorilla/mux"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type V1Dependencies struct {
	Clients *clients.Clients
	Config  *config.Config
}

type PutPing struct {
	Ping string `json:"ping"`
}

func CreateRouter(deps V1Dependencies) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/pings/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		workflowId := vars["id"]

		handle := deps.Clients.Temporal.GetWorkflow(r.Context(), workflowId, "")
		var result string
		err := handle.Get(r.Context(), &result)
		if err != nil {
			if _, ok := err.(*serviceerror.NotFound); ok {
				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
	}).Methods("GET")

	router.HandleFunc("/pings/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		workflowId := vars["id"]

		var putPing PutPing
		err := json.NewDecoder(r.Body).Decode(&putPing)
		if err != nil || putPing.Ping == "" {
			http.Error(w, "'ping' body attribute is a required input", http.StatusBadRequest)
			return
		}

		options := client.StartWorkflowOptions{
			ID:                    workflowId,
			TaskQueue:             deps.Config.Temporal.Worker.TaskQueue,
			WorkflowIDReusePolicy: client.WorkflowIDReusePolicyAllowDuplicateFailedOnly,
		}

		we, err := deps.Clients.Temporal.ExecuteWorkflow(r.Context(), options, pingWorkflow, putPing.Ping)
		if err != nil {
			if _, ok := err.(*temporal.WorkflowExecutionAlreadyStartedError); ok {
				http.Error(w, "PingWorkflow '"+workflowId+"' has already been started.", http.StatusConflict)
				return
			}
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", "./"+workflowId)
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(we.GetID())
	}).Methods("PUT")

	return router
}
