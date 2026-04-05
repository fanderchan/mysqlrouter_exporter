package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "/etc/mysqlrouter_exporter/config.yml"

type Config struct {
	ListenAddress               string `yaml:"listen_address"`
	MetricsPath                 string `yaml:"metrics_path"`
	APIBaseURL                  string `yaml:"api_base_url"`
	APIUser                     string `yaml:"api_user"`
	APIPassword                 string `yaml:"api_password"`
	APIPasswordFile             string `yaml:"api_password_file"`
	TimeoutSeconds              int    `yaml:"timeout_seconds"`
	InsecureSkipVerify          bool   `yaml:"insecure_skip_verify"`
	TLSCAFile                   string `yaml:"tls_ca_file"`
	TLSCertFile                 string `yaml:"tls_cert_file"`
	TLSKeyFile                  string `yaml:"tls_key_file"`
	CollectRouteConnections     bool   `yaml:"collect_route_connections"`
	RouterConfigFile            string `yaml:"router_config_file"`
	ListenerCheckEnabled        bool   `yaml:"listener_check_enabled"`
	ListenerCheckTimeoutSeconds int    `yaml:"listener_check_timeout_seconds"`
}

type options struct {
	showVersion bool
	configPath  string
	flagSet     map[string]bool

	listenAddress               string
	listenPort                  int
	metricsPath                 string
	apiBaseURL                  string
	apiBaseURLCompat            string
	apiUser                     string
	apiUserCompat               string
	apiPassword                 string
	apiPasswordCompat           string
	apiPasswordFile             string
	timeoutSeconds              int
	insecureSkipVerify          bool
	tlsCAFile                   string
	tlsCAFileCompat             string
	tlsCertFile                 string
	tlsCertFileCompat           string
	tlsKeyFile                  string
	tlsKeyFileCompat            string
	collectRouteConnections     bool
	routerConfigFile            string
	listenerCheckEnabled        bool
	listenerCheckTimeoutSeconds int
}

func defaultConfig() Config {
	return Config{
		ListenAddress:               ":9165",
		MetricsPath:                 "/metrics",
		APIBaseURL:                  "https://127.0.0.1:8443/api/20190715",
		APIUser:                     "router_api_user",
		APIPassword:                 "",
		APIPasswordFile:             "",
		TimeoutSeconds:              5,
		InsecureSkipVerify:          true,
		TLSCAFile:                   "",
		TLSCertFile:                 "",
		TLSKeyFile:                  "",
		CollectRouteConnections:     false,
		RouterConfigFile:            "/var/lib/mysqlrouter/mysqlrouter.conf",
		ListenerCheckEnabled:        true,
		ListenerCheckTimeoutSeconds: 1,
	}
}

func parseOptions(args []string) (options, error) {
	opts := options{
		flagSet: make(map[string]bool),
	}

	fs := flag.NewFlagSet("mysqlrouter_exporter", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	fs.BoolVar(&opts.showVersion, "version", false, "Print version and exit")
	fs.StringVar(&opts.configPath, "config", "", "Path to configuration file")
	fs.StringVar(&opts.listenAddress, "web.listen-address", "", "Address to listen on for web endpoints")
	fs.IntVar(&opts.listenPort, "listen-port", 0, "Compatibility flag: set listen port only")
	fs.StringVar(&opts.metricsPath, "web.metrics-path", "", "Path under which to expose metrics")
	fs.StringVar(&opts.apiBaseURL, "router.api-base-url", "", "Base URL of the MySQL Router REST API")
	fs.StringVar(&opts.apiBaseURLCompat, "url", "", "Compatibility flag: REST API URL")
	fs.StringVar(&opts.apiUser, "router.api-user", "", "REST API username")
	fs.StringVar(&opts.apiUserCompat, "user", "", "Compatibility flag: REST API username")
	fs.StringVar(&opts.apiPassword, "router.api-password", "", "REST API password")
	fs.StringVar(&opts.apiPasswordCompat, "pass", "", "Compatibility flag: REST API password")
	fs.StringVar(&opts.apiPasswordFile, "router.api-password-file", "", "Path to a file containing the REST API password")
	fs.IntVar(&opts.timeoutSeconds, "router.timeout-seconds", 0, "REST API timeout in seconds")
	fs.BoolVar(&opts.insecureSkipVerify, "router.insecure-skip-verify", false, "Skip TLS certificate verification")
	fs.BoolVar(&opts.insecureSkipVerify, "skip-tls-verify", false, "Compatibility flag: skip TLS certificate verification")
	fs.StringVar(&opts.tlsCAFile, "router.tls-ca-file", "", "CA file used to validate the Router TLS certificate")
	fs.StringVar(&opts.tlsCAFileCompat, "tls-ca-cert-path", "", "Compatibility flag: CA file path")
	fs.StringVar(&opts.tlsCertFile, "router.tls-cert-file", "", "Client certificate file for Router mTLS")
	fs.StringVar(&opts.tlsCertFileCompat, "tls-cert-path", "", "Compatibility flag: client certificate path")
	fs.StringVar(&opts.tlsKeyFile, "router.tls-key-file", "", "Client key file for Router mTLS")
	fs.StringVar(&opts.tlsKeyFileCompat, "tls-key-path", "", "Compatibility flag: client key path")
	fs.BoolVar(&opts.collectRouteConnections, "collector.route-connections", false, "Collect route connection metrics")
	fs.StringVar(&opts.routerConfigFile, "router.config-file", "", "Path to mysqlrouter.conf for listener checks")
	fs.BoolVar(&opts.listenerCheckEnabled, "collector.listener-check", false, "Enable listener checks against mysqlrouter.conf")
	fs.IntVar(&opts.listenerCheckTimeoutSeconds, "collector.listener-check-timeout-seconds", 0, "Listener check timeout in seconds")

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}

	fs.Visit(func(f *flag.Flag) {
		opts.flagSet[f.Name] = true
	})

	return opts, nil
}

func buildConfig(opts options) (Config, string, error) {
	cfg := defaultConfig()
	cfgSource := "defaults"

	configPath, configPathExplicit := resolveConfigPath(opts)
	if configPathExplicit {
		if err := loadConfigFile(configPath, &cfg); err != nil {
			return Config{}, "", err
		}
		cfgSource = configPath
	} else if fileExists(defaultConfigPath) {
		if err := loadConfigFile(defaultConfigPath, &cfg); err != nil {
			return Config{}, "", err
		}
		cfgSource = defaultConfigPath
	}

	applyEnvOverrides(&cfg)
	applyFlagOverrides(&cfg, opts)

	if err := cfg.normalize(); err != nil {
		return Config{}, "", err
	}
	if err := cfg.validate(); err != nil {
		return Config{}, "", err
	}

	return cfg, cfgSource, nil
}

func resolveConfigPath(opts options) (string, bool) {
	if opts.flagSet["config"] {
		return opts.configPath, true
	}

	if value, ok := lookupEnvAny("MYSQLROUTER_EXPORTER_CONFIG"); ok {
		return value, true
	}

	return defaultConfigPath, false
}

func loadConfigFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config %s: %w", path, err)
	}
	return nil
}

func applyEnvOverrides(cfg *Config) {
	applyEnvString(&cfg.ListenAddress, "MYSQLROUTER_EXPORTER_WEB_LISTEN_ADDRESS")
	applyEnvListenPort(cfg, "MYSQLROUTER_EXPORTER_LISTEN_PORT")
	applyEnvString(&cfg.MetricsPath, "MYSQLROUTER_EXPORTER_WEB_METRICS_PATH")
	applyEnvString(&cfg.APIBaseURL, "MYSQLROUTER_EXPORTER_ROUTER_API_BASE_URL", "MYSQLROUTER_EXPORTER_URL")
	applyEnvString(&cfg.APIUser, "MYSQLROUTER_EXPORTER_ROUTER_API_USER", "MYSQLROUTER_EXPORTER_USER")
	applyEnvString(&cfg.APIPassword, "MYSQLROUTER_EXPORTER_ROUTER_API_PASSWORD", "MYSQLROUTER_EXPORTER_PASS")
	applyEnvString(&cfg.APIPasswordFile, "MYSQLROUTER_EXPORTER_ROUTER_API_PASSWORD_FILE")
	applyEnvInt(&cfg.TimeoutSeconds, "MYSQLROUTER_EXPORTER_ROUTER_TIMEOUT_SECONDS")
	applyEnvBool(&cfg.InsecureSkipVerify, "MYSQLROUTER_EXPORTER_ROUTER_INSECURE_SKIP_VERIFY", "MYSQLROUTER_TLS_SKIP_VERIFY")
	applyEnvString(&cfg.TLSCAFile, "MYSQLROUTER_EXPORTER_ROUTER_TLS_CA_FILE", "MYSQLROUTER_TLS_CACERT_PATH")
	applyEnvString(&cfg.TLSCertFile, "MYSQLROUTER_EXPORTER_ROUTER_TLS_CERT_FILE", "MYSQLROUTER_TLS_CERT_PATH")
	applyEnvString(&cfg.TLSKeyFile, "MYSQLROUTER_EXPORTER_ROUTER_TLS_KEY_FILE", "MYSQLROUTER_TLS_KEY_PATH")
	applyEnvBool(&cfg.CollectRouteConnections, "MYSQLROUTER_EXPORTER_COLLECT_ROUTE_CONNECTIONS")
	applyEnvString(&cfg.RouterConfigFile, "MYSQLROUTER_EXPORTER_ROUTER_CONFIG_FILE")
	applyEnvBool(&cfg.ListenerCheckEnabled, "MYSQLROUTER_EXPORTER_LISTENER_CHECK_ENABLED")
	applyEnvInt(&cfg.ListenerCheckTimeoutSeconds, "MYSQLROUTER_EXPORTER_LISTENER_CHECK_TIMEOUT_SECONDS")
}

func applyFlagOverrides(cfg *Config, opts options) {
	if opts.flagSet["web.listen-address"] {
		cfg.ListenAddress = opts.listenAddress
	}
	if opts.flagSet["listen-port"] {
		cfg.ListenAddress = applyListenPort(cfg.ListenAddress, opts.listenPort)
	}
	if opts.flagSet["web.metrics-path"] {
		cfg.MetricsPath = opts.metricsPath
	}
	if opts.flagSet["router.api-base-url"] {
		cfg.APIBaseURL = opts.apiBaseURL
	}
	if opts.flagSet["url"] {
		cfg.APIBaseURL = opts.apiBaseURLCompat
	}
	if opts.flagSet["router.api-user"] {
		cfg.APIUser = opts.apiUser
	}
	if opts.flagSet["user"] {
		cfg.APIUser = opts.apiUserCompat
	}
	if opts.flagSet["router.api-password"] {
		cfg.APIPassword = opts.apiPassword
	}
	if opts.flagSet["pass"] {
		cfg.APIPassword = opts.apiPasswordCompat
	}
	if opts.flagSet["router.api-password-file"] {
		cfg.APIPasswordFile = opts.apiPasswordFile
	}
	if opts.flagSet["router.timeout-seconds"] {
		cfg.TimeoutSeconds = opts.timeoutSeconds
	}
	if opts.flagSet["router.insecure-skip-verify"] || opts.flagSet["skip-tls-verify"] {
		cfg.InsecureSkipVerify = opts.insecureSkipVerify
	}
	if opts.flagSet["router.tls-ca-file"] {
		cfg.TLSCAFile = opts.tlsCAFile
	}
	if opts.flagSet["tls-ca-cert-path"] {
		cfg.TLSCAFile = opts.tlsCAFileCompat
	}
	if opts.flagSet["router.tls-cert-file"] {
		cfg.TLSCertFile = opts.tlsCertFile
	}
	if opts.flagSet["tls-cert-path"] {
		cfg.TLSCertFile = opts.tlsCertFileCompat
	}
	if opts.flagSet["router.tls-key-file"] {
		cfg.TLSKeyFile = opts.tlsKeyFile
	}
	if opts.flagSet["tls-key-path"] {
		cfg.TLSKeyFile = opts.tlsKeyFileCompat
	}
	if opts.flagSet["collector.route-connections"] {
		cfg.CollectRouteConnections = opts.collectRouteConnections
	}
	if opts.flagSet["router.config-file"] {
		cfg.RouterConfigFile = opts.routerConfigFile
	}
	if opts.flagSet["collector.listener-check"] {
		cfg.ListenerCheckEnabled = opts.listenerCheckEnabled
	}
	if opts.flagSet["collector.listener-check-timeout-seconds"] {
		cfg.ListenerCheckTimeoutSeconds = opts.listenerCheckTimeoutSeconds
	}
}

func (c *Config) normalize() error {
	c.APIBaseURL = strings.TrimRight(c.APIBaseURL, "/")
	if c.MetricsPath == "" {
		c.MetricsPath = "/metrics"
	} else if !strings.HasPrefix(c.MetricsPath, "/") {
		c.MetricsPath = "/" + c.MetricsPath
	}
	if c.ListenAddress == "" {
		c.ListenAddress = ":9165"
	} else if _, _, err := net.SplitHostPort(c.ListenAddress); err != nil && !strings.HasPrefix(c.ListenAddress, ":") {
		if _, err := strconv.Atoi(c.ListenAddress); err == nil {
			c.ListenAddress = ":" + c.ListenAddress
		}
	}
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = 5
	}
	if c.RouterConfigFile == "" {
		c.RouterConfigFile = "/var/lib/mysqlrouter/mysqlrouter.conf"
	}
	if c.ListenerCheckTimeoutSeconds <= 0 {
		c.ListenerCheckTimeoutSeconds = 1
	}
	if c.APIPassword == "" && c.APIPasswordFile != "" {
		password, err := os.ReadFile(c.APIPasswordFile)
		if err != nil {
			return fmt.Errorf("read api_password_file %s: %w", c.APIPasswordFile, err)
		}
		c.APIPassword = strings.TrimSpace(string(password))
	}
	return nil
}

func (c Config) validate() error {
	if c.APIBaseURL == "" {
		return fmt.Errorf("api_base_url is required")
	}
	if c.APIUser == "" {
		return fmt.Errorf("api_user is required")
	}
	if c.APIPassword == "" {
		return fmt.Errorf("api_password or api_password_file is required")
	}
	if (c.TLSCertFile == "") != (c.TLSKeyFile == "") {
		return fmt.Errorf("tls_cert_file and tls_key_file must be set together")
	}
	return nil
}

func lookupEnvAny(keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			return value, true
		}
	}
	return "", false
}

func applyEnvString(target *string, keys ...string) {
	if value, ok := lookupEnvAny(keys...); ok {
		*target = value
	}
}

func applyEnvBool(target *bool, keys ...string) {
	value, ok := lookupEnvAny(keys...)
	if !ok {
		return
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err == nil {
		*target = parsed
	}
}

func applyEnvInt(target *int, keys ...string) {
	value, ok := lookupEnvAny(keys...)
	if !ok {
		return
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err == nil {
		*target = parsed
	}
}

func applyEnvListenPort(cfg *Config, keys ...string) {
	value, ok := lookupEnvAny(keys...)
	if !ok {
		return
	}

	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return
	}
	cfg.ListenAddress = applyListenPort(cfg.ListenAddress, port)
}

func applyListenPort(address string, port int) string {
	if port <= 0 {
		return address
	}

	if host, _, err := net.SplitHostPort(address); err == nil {
		return net.JoinHostPort(host, strconv.Itoa(port))
	}

	if strings.HasPrefix(address, ":") || address == "" {
		return ":" + strconv.Itoa(port)
	}

	return net.JoinHostPort(address, strconv.Itoa(port))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
