package collector

import (
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "cmc"
)

var (
	defaultCoinStatsLabels      = []string{"symbol"}
	defaultCoinStatsLabelValues = func(coinSymbol string) []string {
		return []string{coinSymbol}
	}
)

//Struct for scaping coinmarket data
type Coin struct {
	Rank               float64
	Name               string
	Symbol             string
	Market_Cap         float64
	Price              float64
	Circulating_Supply float64
	Volume_24h         float64
	Change_1h          float64
	Change_24h         float64
	Change_7d          float64
}

type coinMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(coin Coin) float64
	Labels func(coinSymbol string) []string
}

type CoinStats struct {
	logger log.Logger
	client *http.Client
	url    *url.URL

	up                              prometheus.Gauge
	totalScrapes, htmlParseFailures prometheus.Counter

	metrics []*coinMetric
}

// NewCoinStats returns a new Collector exposing Coin stats.
func NewCoinStats(logger log.Logger, client *http.Client, url *url.URL) *CoinStats {
	subsystem := "coin_stats"

	return &CoinStats{
		logger: logger,
		client: client,
		url:    url,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "up"),
			Help: "Was the last scrape of the website successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "total_scrapes"),
			Help: "Current total website scrapes.",
		}),
		htmlParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "html_parse_failures"),
			Help: "Number of errors while parsing HTML.",
		}),

		metrics: []*coinMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "market_cap"),
					"The total market value of a cryptocurrency's circulating supply.",
					defaultCoinStatsLabels, nil,
				),
				Value: func(coinStat Coin) float64 {
					return float64(coinStat.Market_Cap)
				},
				Labels: defaultCoinStatsLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "price"),
					"Current price in USD.",
					defaultCoinStatsLabels, nil,
				),
				Value: func(coinStat Coin) float64 {
					return float64(coinStat.Price)
				},
				Labels: defaultCoinStatsLabelValues,
			},
		},
	}
}

// Describe set Prometheus metrics descriptions.
func (c *CoinStats) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Desc
	}

	ch <- c.up.Desc()
	ch <- c.totalScrapes.Desc()
	ch <- c.htmlParseFailures.Desc()
}

func (c *CoinStats) fetchAndDecodeCoinStats() ([]Coin, error) {
	coins := make([]Coin, 0)

	var coin Coin
	coin.Price = 100
	coin.Symbol = "BTC"
	coins = append(coins, coin)

	var coin1 Coin
	coin1.Price = 50
	coin1.Symbol = "ETH"
	coins = append(coins, coin1)

	return coins, nil
}

// Collect collects CoinStats metrics.
func (c *CoinStats) Collect(ch chan<- prometheus.Metric) {
	var err error
	c.totalScrapes.Inc()
	defer func() {
		ch <- c.up
		ch <- c.totalScrapes
		ch <- c.htmlParseFailures
	}()

	coinStatsResp, err := c.fetchAndDecodeCoinStats()
	if err != nil {
		c.up.Set(0)
		_ = level.Warn(c.logger).Log(
			"msg", "failed to fetch and decode coin stats",
			"err", err,
		)
		return
	}
	c.up.Set(1)

	for _, coinStats := range coinStatsResp {
		for _, metric := range c.metrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(coinStats),
				metric.Labels(coinStats.Symbol)...,
			)
		}
	}
}
