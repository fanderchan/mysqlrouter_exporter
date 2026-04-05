package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	opts, err := parseOptions(os.Args[1:])
	if err != nil {
		log.Printf("failed to parse options: %v", err)
		os.Exit(2)
	}

	if opts.showVersion {
		fmt.Println(version)
		return
	}

	cfg, cfgSource, err := buildConfig(opts)
	if err != nil {
		log.Printf("failed to build config: %v", err)
		os.Exit(1)
	}

	exporter, err := NewExporter(cfg, BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    buildDate,
	})
	if err != nil {
		log.Printf("failed to initialize exporter: %v", err)
		os.Exit(1)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready\n"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(
			w,
			"mysqlrouter_exporter is running.\nmetrics: %s\nhealth: /-/healthy\nready: /-/ready\n",
			cfg.MetricsPath,
		)
	})

	server := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf(
		"mysqlrouter_exporter starting version=%s commit=%s build_date=%s config_source=%s listen=%s metrics_path=%s",
		version,
		commit,
		buildDate,
		cfgSource,
		cfg.ListenAddress,
		cfg.MetricsPath,
	)

	if err := server.ListenAndServe(); err != nil {
		log.Printf("http server stopped: %v", err)
		os.Exit(1)
	}
}
