// Command proxy is the entry-point for the privacy proxy engine.
//
// Usage:
//
//	proxy [flags]
//
// Flags:
//
//	-addr       TCP address to listen on (default ":8080")
//	-metrics    TCP address for the Prometheus metrics endpoint (default ":9090")
//	-config     Path to YAML configuration file (default "configs/proxy.yaml")
//	-hops       Number of anonymisation circuit hops (default 3)
package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Cipher7788/media-hub/proxy/internal/anonymize"
	"github.com/Cipher7788/media-hub/proxy/internal/core"
	"github.com/Cipher7788/media-hub/proxy/internal/firewall"
	"github.com/Cipher7788/media-hub/proxy/internal/monitor"
	"github.com/Cipher7788/media-hub/proxy/internal/tracking"
)

func main() {
	addr := flag.String("addr", ":8080", "proxy listen address")
	metricsAddr := flag.String("metrics", ":9090", "metrics listen address")
	hops := flag.Int("hops", 3, "anonymisation circuit hops")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ---- Firewall & DPI ----
	fw := firewall.New(firewall.Allow)
	// Block private / RFC-1918 ranges to prevent SSRF.
	for _, cidr := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8"} {
		rule, err := firewall.BlockCIDR(cidr)
		if err != nil {
			logger.Error("invalid CIDR", "cidr", cidr, "err", err)
			os.Exit(1)
		}
		fw.AddRule(rule)
	}

	dpi := firewall.NewInspector(firewall.ProtocolBitTorrent, firewall.ProtocolP2P)

	// ---- Anonymisation ----
	sanitiser := anonymize.DefaultSanitiser()

	cb := anonymize.NewCircuitBuilder(*hops)
	// In production these come from a configuration file or discovery service.
	for i := 1; i <= *hops+1; i++ {
		cb.RegisterNode(anonymize.Node{
			Addr:   "127.0.0.1:0", // placeholder; replace with real relay addresses
			Weight: 1,
		})
	}

	// ---- Tracking Prevention ----
	trackingEngine := tracking.New()

	// ---- Monitoring ----
	metrics := monitor.NewMetrics()
	detector := monitor.NewDetector(
		time.Minute,
		300, // 300 req/min threshold
		func(a monitor.Alert) {
			logger.Warn("threat alert",
				"level", a.Level.String(),
				"source", a.Source,
				"message", a.Message,
			)
		},
		logger,
	)

	// ---- Proxy ----
	cfg := core.DefaultConfig()
	cfg.ListenAddr = *addr

	p := core.New(cfg, logger)
	p.Use(fw.Middleware())
	p.Use(dpi.Middleware())
	p.Use(sanitiser.Middleware())
	p.Use(trackingEngine.Middleware())
	p.Use(detector.Middleware())

	// Log circuit info at startup.
	circuit, err := cb.Build()
	if err != nil {
		logger.Warn("could not build anonymisation circuit", "err", err)
	} else {
		logger.Info("anonymisation circuit ready",
			"hops", circuit.Len(),
		)
	}

	// Start the metrics server in the background.
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", metrics.Handler())
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		logger.Info("metrics server listening", "addr", *metricsAddr)
		if err := http.ListenAndServe(*metricsAddr, mux); err != nil {
			logger.Error("metrics server error", "err", err)
		}
	}()

	// Block until the proxy exits.
	if err := p.ListenAndServe(); err != nil {
		logger.Error("proxy exited", "err", err)
		os.Exit(1)
	}
}
