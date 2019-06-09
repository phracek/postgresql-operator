package k8shandler

const (
	postgresqlPort = 5432

	defaultPgImage            = "mcyprian/postgresql-10-fedora29"
	defaultCntCommand         = "statefulset-startup"
	defaultHealthCheckCommand = "/usr/libexec/check-container"

	defaultCPULimit      = "4000m"
	defaultCPURequest    = "100m"
	defaultMemoryLimit   = "4Gi"
	defaultMemoryRequest = "1Gi"
)
