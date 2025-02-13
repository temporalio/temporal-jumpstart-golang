package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Temporal     *TemporalConfig
	IsProduction bool
	API          *APIConfig
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

type TemporalWorkerTuner struct {
	TargetMem float64
	TargetCPU float64
}

type TemporalWorkerRateLimits struct {
	MaxWorkerActivitiesPerSecond    int
	MaxTaskQueueActivitiesPerSecond int
}

type TemporalWorker struct {
	TaskQueue  string
	Name       string
	Capacity   *TemporalWorkerCapacity
	RateLimits *TemporalWorkerRateLimits
	Tuner      *TemporalWorkerTuner
}

type TemporalConnection struct {
	Namespace string
	Target    string
	MTLS      *MTLSConfig
}

type TemporalConfig struct {
	Connection *TemporalConnection
	Worker     *TemporalWorker
}

type APIConfig struct {
	Port string
	MTLS *MTLSConfig
	URL  *url.URL
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize config: %v", err))
	}
	return cfg
}
func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	apiCfg, err := createApiCfg()
	if err != nil {
		return nil, err
	}

	temporalCfg, err := createTemporalCfg()
	if err != nil {
		return nil, err
	}

	return &Config{
		Temporal:     temporalCfg,
		IsProduction: strings.ToLower(os.Getenv("ENV")) == "production",
		API:          apiCfg,
	}, nil
}

func createApiCfg() (*APIConfig, error) {
	apiUrl := os.Getenv("API_URL")
	if apiUrl == "" {
		apiUrl = "https://localhost:4000/api"
	}
	parsedUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}
	mtls := &MTLSConfig{
		CertChainFile:               os.Getenv("API_CONNECTION_MTLS_CERT_CHAIN_FILE"),
		KeyFile:                     os.Getenv("API_CONNECTION_MTLS_KEY_FILE"),
		PKCS:                        os.Getenv("API_CONNECTION_MTLS_PKCS"),
		InsecureTrustManager:        os.Getenv("API_CONNECTION_MTLS_INSECURE_TRUST_MANAGER") == "true",
		KeyPassword:                 os.Getenv("API_CONNECTION_MTLS_KEY_PASSWORD"),
		ServerName:                  os.Getenv("API_CONNECTION_MTLS_SERVER_NAME"),
		ServerRootCACertificateFile: os.Getenv("API_CONNECTION_MTLS_SERVER_ROOT_CA_CERTIFICATE_FILE"),
	}

	if strings.HasPrefix(apiUrl, "https") && (mtls.CertChainFile == "" || mtls.KeyFile == "") {
		return nil, errors.New("Invalid config: HTTPS requires API_CONNECTION_MTLS* settings")
	}

	return &APIConfig{
		Port: parsedUrl.Port(),
		MTLS: mtls,
		URL:  parsedUrl,
	}, nil
}

func createTemporalCfg() (*TemporalConfig, error) {
	mtls := &MTLSConfig{
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

	worker := &TemporalWorker{
		Capacity: &TemporalWorkerCapacity{
			MaxConcurrentWorkflowTaskExecutions: intOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CONCURRENT_WORKFLOW_TASK_EXECUTORS"),
			MaxConcurrentActivityExecutors:      intOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CONCURRENT_ACTIVITY_EXECUTORS"),
			MaxConcurrentLocalActivityExecutors: intOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CONCURRENT_LOCAL_ACTIVITY_EXECUTORS"),
			MaxConcurrentWorkflowTaskPollers:    intOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CONCURRENT_WORKFLOW_TASK_POLLERS"),
			MaxConcurrentActivityTaskPollers:    intOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CONCURRENT_ACTIVITY_TASK_POLLERS"),
			MaxCachedWorkflows:                  intOrNot("TEMPORAL_WORKER_CAPACITY_MAX_CACHED_WORKFLOWS"),
		},
		Name: "",
		RateLimits: &TemporalWorkerRateLimits{
			MaxWorkerActivitiesPerSecond:    intOrNot("TEMPORAL_WORKER_RATE_LIMITS_MAX_WORKER_ACTIVITIES_PER_SECOND"),
			MaxTaskQueueActivitiesPerSecond: intOrNot("TEMPORAL_WORKER_RATE_LIMITS_MAX_TASK_QUEUE_ACTIVITIES_PER_SECOND"),
		},
		TaskQueue: assertCfg("TEMPORAL_WORKER_TASK_QUEUE"),
		Tuner: &TemporalWorkerTuner{
			TargetCPU: floatOrNot("TEMPORAL_WORKER_TUNER_TARGET_CPU"),
			TargetMem: floatOrNot("TEMPORAL_WORKER_TUNER_TARGET_MEM"),
		},
	}

	connection := &TemporalConnection{
		Namespace: assertCfg("TEMPORAL_CONNECTION_NAMESPACE"),
		Target:    assertCfg("TEMPORAL_CONNECTION_TARGET"),
		MTLS:      mtls,
	}

	return &TemporalConfig{
		Connection: connection,
		Worker:     worker,
	}, nil
}
func floatOrNot(key string) float64 {
	if value := os.Getenv(key); value != "" {
		num, _ := strconv.ParseFloat(value, 64)
		return num
	}
	return 0
}

func intOrNot(key string) int {
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
