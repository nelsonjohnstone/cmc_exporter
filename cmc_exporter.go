package main

import (
	"net/http"
	"net/url"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/nelsonjohnstone/cmc_scraper/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		Name          = "cmc_exporter"
		webConfig     = webflag.AddFlags(kingpin.CommandLine)
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9101").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		cmcScrapeURI  = kingpin.Flag("cmc.scrape-uri", "URI on which to scrape CMC.").Default("https://coinmarketcap.com/all/views/all/").String()
		// cmcSSLVerify  = kingpin.Flag("cmc.ssl-verify", "Flag that enables SSL certificate verification for the scrape URI").Default("true").Bool()
		cmcTimeout = kingpin.Flag("cmc.timeout", "Timeout for trying to get stats from CMC.").Default("5s").Duration()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print(Name))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting cmc_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	cmcUrl, err := url.Parse(*cmcScrapeURI)
	if err != nil {
		_ = level.Error(logger).Log(
			"msg", "failed to parse es.uri",
			"err", err,
		)
		os.Exit(1)
	}

	httpClient := &http.Client{
		Timeout: *cmcTimeout,
		// Transport: httpTransport,
	}

	// version metric
	versionMetric := version.NewCollector(Name)
	prometheus.MustRegister(versionMetric)
	prometheus.MustRegister(collector.NewCoinStats(logger, httpClient, cmcUrl))

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>CMC Exporter</title></head>
             <body>
             <h1>CMC Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	srv := &http.Server{Addr: *listenAddress}
	if err := web.ListenAndServe(srv, *webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
