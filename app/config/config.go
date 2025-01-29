package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Temporal     TemporalConfig
	IsProduction bool
	API          APIConfig
}

type MTLSConfig struct {
	PKCS                        string
	Key                         []byte
	KeyFile                     string
	CertChain                   []byte
	CertChainFile               string
	KeyPassword                 string
	InsecureTrustManager        bool
	ServerName                  string
	ServerRootCACertificate     []byte
	ServerRootCACertificateFile string
}

type TemporalWorkerCapacity struct {
	MaxConcurrentWorkflowTaskExecutions int
	MaxConcurrentActivityExecutors      int
	MaxConcurrentLocalActivityExecutors int
	MaxConcurrentWorkflowTaskPollers    int
	MaxConcurrentActivityTaskPollers    int
	MaxCachedWorkflows                  int
}

type TemporalWorkerRateLimits struct {
	MaxWorkerActivitiesPerSecond    int
	MaxTaskQueueActivitiesPerSecond int
}

type TemporalWorker struct {
	TaskQueue  string
	Name       string
	Capacity   TemporalWorkerCapacity
	RateLimits TemporalWorkerRateLimits
	BundlePath string
}

type TemporalConnection struct {
	Namespace string
	Target    string
	MTLS      MTLSConfig
}

type TemporalConfig struct {
	Connection TemporalConnection
	Worker     TemporalWorker
}

type APIConfig struct {
	Port string
	MTLS MTLSConfig
	URL  string
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}

	apiCfg, err := createApiCfg()
	if err != nil {
		return Config{}, err
	}

	temporalCfg, err := createTemporalCfg()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Temporal:     temporalCfg,
		IsProduction: strings.ToLower(os.Getenv("ENV")) == "production",
		API:          apiCfg,
	}, nil
}

func createApiCfg() (APIConfig, error) {
	apiUrl := os.Getenv("API_URL")
	if apiUrl == "" {
		apiUrl = "https://localhost:4000/api"
	}

	mtls := MTLSConfig{
		CertChainFile:               os.Getenv("API_CONNECTION_MTLS_CERT_CHAIN_FILE"),
		KeyFile:                     os.Getenv("API_CONNECTION_MTLS_KEY_FILE"),
		PKCS:                        os.Getenv("API_CONNECTION_MTLS_PKCS"),
		InsecureTrustManager:        os.Getenv("API_CONNECTION_MTLS_INSECURE_TRUST_MANAGER") == "true",
		KeyPassword:                 os.Getenv("API_CONNECTION_MTLS_KEY_PASSWORD"),
		ServerName:                  os.Getenv("API_CONNECTION_MTLS_SERVER_NAME"),
		ServerRootCACertificateFile: os.Getenv("API_CONNECTION_MTLS_SERVER_ROOT_CA_CERTIFICATE_FILE"),
	}

	if strings.HasPrefix(apiUrl, "https") && (mtls.CertChainFile == "" || mtls.KeyFile == "") {
		return APIConfig{}, errors.New("Invalid config: HTTPS requires API_CONNECTION_MTLS* settings")
	}

	return APIConfig{
		Port: apiUrl,
		MTLS: mtls,
		URL:  apiUrl,
	}, nil
}

func createTemporalCfg() (TemporalConfig, error) {
	mtls := MTLSConfig{
		CertChainFile:               os.Getenv("TEMPORAL_CONNECTION_MTLS_CERT_CHAIN_FILE"),
		KeyFile:                     os.Getenv("TEMPORAL_CONNECTION_MTLS_KEY_FILE"),
		PKCS:                        os.Getenv("TEMPORAL_CONNECTION_MTLS_PKCS"),
		InsecureTrustManager:        os.Getenv("TEMPORAL_CONNECTION_MTLS_INSECURE_TRUST_MANAGER") == "true",
		KeyPassword:                 os.Getenv("TEMPORAL_CONNECTION_MTLS_KEY_PASSWORD"),
		ServerName:                  os.Getenv("TEMPORAL_CONNECTION_MTLS_SERVER_NAME"),
		ServerRootCACertificateFile: os.Getenv("TEMPORAL_CONNECTION_MTLS_SERVER_ROOT_CA_CERTIFICATE_FILE"),
	}

	if key := os.Getenv("TEMPORAL_CONNECTION_MTLS_KEY"); key != "" {
		mtls.Key = []byte(key)
	} else if mtls.KeyFile != "" {
		mtls.Key, _ = os.ReadFile(mtls.KeyFile)
	}

	if certChain := os.Getenv("TEMPORAL_CONNECTION_MTLS_CERT_CHAIN"); certChain != "" {
		mtls.CertChain = []byte(certChain)
	} else if mtls.CertChainFile != "" {
		mtls.CertChain, _ = os.ReadFile(mtls.CertChainFile)
	}

	if rootCACert := os.Getenv("TEMPORAL_CONNECTION_MTLS_SERVER_ROOT_CA_CERTIFICATE"); rootCACert != "" {
		mtls.ServerRootCACertificate = []byte(rootCACert)
	} else if mtls.ServerRootCACertificateFile != "" {
		mtls.ServerRootCACertificate, _ = os.ReadFile(mtls.ServerRootCACertificateFile)
	}

	worker := TemporalWorker{
		Capacity: TemporalWorkerCapacity{
			MaxConcurrentWorkflowTaskExecutions: numOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CONCURRENT_WORKFLOW_TASK_EXECUTORS"),
			MaxConcurrentActivityExecutors:      numOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CONCURRENT_ACTIVITY_EXECUTORS"),
			MaxConcurrentLocalActivityExecutors: numOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CONCURRENT_LOCAL_ACTIVITY_EXECUTORS"),
			MaxConcurrentWorkflowTaskPollers:    numOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CONCURRENT_WORKFLOW_TASK_POLLERS"),
			MaxConcurrentActivityTaskPollers:    numOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CONCURRENT_ACTIVITY_TASK_POLLERS"),
			MaxCachedWorkflows:                  numOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CACHED_WORKFLOWS"),
		},
		Name: "",
		RateLimits: TemporalWorkerRateLimits{
			MaxWorkerActivitiesPerSecond:    numOrNot("TEMPORAL_WORKER_RATE_LIMITS_MAX_WORKER_ACTIVITIES_PER_SECOND"),
			MaxTaskQueueActivitiesPerSecond: numOrNot("TEMPORAL_WORKER_RATE_LIMITS_MAX_TASK_QUEUE_ACTIVITIES_PER_SECOND"),
		},
		TaskQueue:  assertCfg("TEMPORAL_WORKER_TASK_QUEUE"),
		BundlePath: assertCfg("TEMPORAL_WORKER_BUNDLE_PATH"),
	}

	connection := TemporalConnection{
		Namespace: assertCfg("TEMPORAL_CONNECTION_NAMESPACE"),
		Target:    assertCfg("TEMPORAL_CONNECTION_TARGET"),
		MTLS:      mtls,
	}

	return TemporalConfig{
		Connection: connection,
		Worker:     worker,
	}, nil
}

func numOrNot(key string) int {
	if value := os.Getenv(key); value != "" {
		num, _ := strconv.Atoi(value)
		return num
	}
	return 0
}

func assertCfg(name string) string {
	value := os.Getenv(name)
	if value == "" {
		panic(name + " environment variable is not defined")
	}
	return value
}
