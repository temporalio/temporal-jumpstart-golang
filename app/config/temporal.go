package config

type TemporalConfig struct {
	Namespaces map[string]*NamespaceConfig `yaml:"namespaces"`
	Prometheus *PrometheusConfig           `yaml:"prometheus"`
}

type NamespaceConfig struct {
	Connection *ConnectionConfig           `yaml:"connection"`
	TaskQueues map[string]*TaskQueueConfig `yaml:"task_queues"`
	Workflows  *WorkflowsConfig            `yaml:"workflows"`
}
type WorkflowsConfig struct {
	ClientCount int `yaml:"client_count"`
}
type ConnectionConfig struct {
	Target string      `yaml:"target"`
	APIKey string      `yaml:"api_key"`
	MTLS   *MTLSConfig `yaml:"mtls"`
}

type MTLSConfig struct {
	CertChainFile               string `yaml:"cert_chain_file"`
	KeyFile                     string `yaml:"key_file"`
	KeyPassword                 string `yaml:"key_password"`
	InsecureTrustManager        bool   `yaml:"insecure_trust_manager"`
	ServerName                  string `yaml:"server_name"`
	ServerRootCACertificateFile string `yaml:"server_root_ca_certificate_file"`
}

type TaskQueueConfig struct {
	WorkerCount        int               `yaml:"worker_count"`
	MaxCachedWorkflows int               `yaml:"max_cached_workflows"`
	Capacity           *CapacityConfig   `yaml:"capacity"`
	RateLimits         *RateLimitsConfig `yaml:"rate_limits"`
	Tuner              *TunerConfig      `yaml:"tuner"`
}

type CapacityConfig struct {
	MaxConcurrentWorkflowTaskPollers    int `yaml:"max_concurrent_workflow_task_pollers"`
	MaxConcurrentWorkflowTaskExecutors  int `yaml:"max_concurrent_workflow_task_executors"`
	MaxConcurrentActivityTaskPollers    int `yaml:"max_concurrent_activity_task_pollers"`
	MaxConcurrentActivityTaskExecutors  int `yaml:"max_concurrent_activity_task_executors"`
	MaxConcurrentLocalActivityExecutors int `yaml:"max_concurrent_local_activity_executors"`
}

type RateLimitsConfig struct {
	MaxWorkerActivitiesPerSecond    int `yaml:"max_worker_activities_per_second"`
	MaxTaskQueueActivitiesPerSecond int `yaml:"max_task_queue_activities_per_second"`
}

type TunerConfig struct {
	TargetMem float64 `yaml:"target_mem"`
	TargetCPU float64 `yaml:"target_cpu"`
}

type PrometheusConfig struct {
	Address string `yaml:"address"`
	Prefix  string `yaml:"prefix"`
}
