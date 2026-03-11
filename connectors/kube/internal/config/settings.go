package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"common/utils"
)

type Settings struct {
	Name       string `json:"connector_name"                env:"NAME"`
	ListenHost string `json:"listen_host"                   env:"LISTEN_HOST"`
	ListenPort int    `json:"listen_port"                   env:"LISTEN_PORT"`

	InCluster      bool   `json:"in_cluster"                    env:"IN_CLUSTER"`
	KubeConfigPath string `json:"kube_config_path"              env:"KUBE_CONFIG_PATH"`

	OtelServiceName      string `json:"otel_service_name"             env:"OTEL_SERVICE_NAME"`
	OtelServiceNamespace string `json:"otel_service_namespace"        env:"OTEL_SERVICE_NAMESPACE"`
	OtelServiceVersion   string `json:"otel_service_version"          env:"OTEL_SERVICE_VERSION"`
	OtlpEndpointGrpc     string `json:"otlp_endpoint_grpc"            env:"OTLP_ENDPOINT_GRPC"`

	LogLevel  string `json:"log_level"                     env:"LOG_LEVEL"`
	LogFormat string `json:"log_format"                    env:"LOG_FORMAT"`

	HealthCheckIntervalSeconds int `json:"health_check_interval_seconds" env:"HEALTH_CHECK_INTERVAL_SECONDS"`
	HealthRetryIntervalSeconds int `json:"health_retry_interval_seconds" env:"HEALTH_RETRY_INTERVAL_SECONDS"`

	Healthy         *bool
	HealthJobID     *string
	HealthNextCheck *string
	HealthLastCheck *string
}

var defaultSettings = Settings{
	Name:       "connector-kube",
	ListenHost: "192.168.1.170",
	ListenPort: 8080,

	InCluster:      false,
	KubeConfigPath: "/home/martin/.kube/config",

	OtelServiceName:      "connector-kube",
	OtelServiceNamespace: "fetch-api",
	OtelServiceVersion:   "",
	OtlpEndpointGrpc:     "grpc.k8s.iaminyourpc.xyz:80",

	LogLevel:  "info",
	LogFormat: "logfmt",

	HealthCheckIntervalSeconds: 15,
	HealthRetryIntervalSeconds: 5,

	Healthy:     nil,
	HealthJobID: nil,
}

var Config Settings

func init() {
	config, err := loadSettings()

	if err != nil {
		panic(err)
	}

	Config = config
}

func loadSettings() (Settings, error) {
	settings := Settings{}
	defaults := reflect.ValueOf(defaultSettings)

	fields := reflect.ValueOf(&settings).Elem()
	fieldsMeta := fields.Type()

	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		fieldMeta := fieldsMeta.Field(i)

		envName := fieldMeta.Tag.Get("env")
		if envName == "" {
			continue
		}

		envVal, envDefined := os.LookupEnv(envName)
		if !envDefined {
			field.Set(defaults.Field(i))

			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(envVal)

		case reflect.Int, reflect.Int64:
			val, err := utils.ToInt(envVal)

			if err != nil {
				return Settings{}, fmt.Errorf("%s got an incorrect value: %s", envName, err)
			}

			field.SetInt(val)

		case reflect.Bool:
			val, err := utils.ToBool(envVal)

			if err != nil {
				return Settings{}, fmt.Errorf("%s got an incorrect value: %s", envName, err)
			}

			field.SetBool(val)
		}
	}

	if !settings.InCluster && settings.KubeConfigPath == "" {
		return Settings{}, fmt.Errorf("KUBE_CONFIG_PATH is required when IN_CLUSTER is set to false")
	}

	if settings.OtelServiceVersion == "" {
		_, currentFile, _, ok := runtime.Caller(0)

		if !ok {
			settings.OtelServiceVersion = "unknown"
		} else {
			currentDir := filepath.Dir(currentFile)
			verFile := filepath.Join(currentDir, "..", "..", "VERSION")

			if appVer, err := utils.ReadFile(verFile); err != nil {
				settings.OtelServiceVersion = "unknown"
			} else {
				settings.OtelServiceVersion = strings.TrimSpace(appVer)
			}
		}
	}

	return settings, nil
}
