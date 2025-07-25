package main

import (
	"context"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/instrumentation/prometheus"
	"github.com/uber-go/tally/v4"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	worker "github.com/temporalio/temporal-jumpstart-golang/app/workers/temporal"
)

type WorkerArgs struct {
	NamespacedTaskQueues []string // namespace:taskqueue pairs
	Env                  string
	ConfigDir            string
}

var rootCmd = &cobra.Command{
	Use:   "workers",
	Short: "Temporal workers CLI",
	Long:  "A CLI tool for managing Temporal workers",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the workers",
	Long:  "Start and run the Temporal workers",
	Run: func(cmd *cobra.Command, args []string) {
		namespacedTaskQueues, _ := cmd.Flags().GetStringSlice("namespaced-task-queue")
		env, _ := cmd.Flags().GetString("environment")
		configDir, _ := cmd.Flags().GetString("config-dir")

		workerArgs := WorkerArgs{
			NamespacedTaskQueues: namespacedTaskQueues,
			Env:                  env,
			ConfigDir:            configDir,
		}

		runWorkers(workerArgs)
	},
}

func init() {
	runCmd.Flags().StringSliceP("namespaced-task-queue", "n", []string{"default:default"}, "Namespaced task queues in format namespace:taskqueue (can specify multiple)")
	runCmd.Flags().StringP("environment", "e", "default", "Environment")
	runCmd.Flags().StringP("config-dir", "c", "config", "Directory where config files are found")

	rootCmd.AddCommand(runCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseWorkerTargets(namespacedTaskQueues []string) ([]worker.WorkerTarget, error) {
	targets := make([]worker.WorkerTarget, 0, len(namespacedTaskQueues))

	for _, ntq := range namespacedTaskQueues {
		parts := strings.Split(ntq, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid namespaced-task-queue format '%s', expected 'namespace:taskqueue'", ntq)
		}

		targets = append(targets, worker.WorkerTarget{
			Namespace: strings.TrimSpace(parts[0]),
			TaskQueue: strings.TrimSpace(parts[1]),
		})
	}

	return targets, nil
}

func runWorkers(args WorkerArgs) {
	setupLogging()

	cfg := config.MustNewConfig(args.ConfigDir, args.Env)
	ctx := context.Background()

	c := clients.MustNewClients(ctx, cfg)

	targets, err := parseWorkerTargets(args.NamespacedTaskQueues)
	if err != nil {
		log.Fatalf("Failed to parse worker targets: %v", err)
	}

	slog.Info("Starting workers", "targets", targets)
	rootScope := setupObservability(err, ctx, cfg)

	workerClients, err := worker.NewClients(ctx, cfg, rootScope, targets)
	if err != nil {
		log.Fatalf("Failed to connect worker clients: %v", err)
	}

	host := worker.NewWorkerHost(
		ctx,
		&worker.HostConfig{
			ShutdownTimeout: 15 * time.Second,
			WorkerType:      "workers",
		},
		&NullBuilder{Clients: c},
	)

	if err = host.Run(ctx, cfg, workerClients...); err != nil {
		slog.Error("Worker host failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Workers completed")
}

func setupObservability(err error, ctx context.Context, cfg *config.Config) tally.Scope {
	reporter, err := prometheus.NewReporter(ctx, cfg,
		prometheus.WithListenAddress(cfg.Temporal.Prometheus.Address))
	if err != nil {
		log.Fatalf("Failed to create prometheus reporter: %v", err)
	}

	rootScope, err := prometheus.NewRootScope(ctx, cfg, prometheus.WithReporter(reporter))
	if err != nil {
		log.Fatalf("Failed to create prometheus root scope: %v", err)
	}
	return rootScope
}

func setupLogging() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	handler := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(handler))
}
