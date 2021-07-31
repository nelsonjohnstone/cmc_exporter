package collector

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gocolly/colly"
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
	url    string

	up                              prometheus.Gauge
	totalScrapes, htmlParseFailures prometheus.Counter

	metrics []*coinMetric
}

// NewCoinStats returns a new Collector exposing Coin stats.
func NewCoinStats(logger log.Logger, client *http.Client, url string) *CoinStats {
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
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "circulating_supply"),
					"The amount of coins that are circulating in the market and are in public hands.",
					defaultCoinStatsLabels, nil,
				),
				Value: func(coinStat Coin) float64 {
					return float64(coinStat.Circulating_Supply)
				},
				Labels: defaultCoinStatsLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "volume_24h"),
					"A measure of how much of a cryptocurrency was traded in the last 24 hours.",
					defaultCoinStatsLabels, nil,
				),
				Value: func(coinStat Coin) float64 {
					return float64(coinStat.Volume_24h)
				},
				Labels: defaultCoinStatsLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "change_1h"),
					"Percentage change in price over the last 1 hour.",
					defaultCoinStatsLabels, nil,
				),
				Value: func(coinStat Coin) float64 {
					return float64(coinStat.Change_1h)
				},
				Labels: defaultCoinStatsLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "change_24h"),
					"Percentage change in price over the last 24 hours.",
					defaultCoinStatsLabels, nil,
				),
				Value: func(coinStat Coin) float64 {
					return float64(coinStat.Change_24h)
				},
				Labels: defaultCoinStatsLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "change_7d"),
					"Percentage change in price over the last 7 days.",
					defaultCoinStatsLabels, nil,
				),
				Value: func(coinStat Coin) float64 {
					return float64(coinStat.Change_7d)
				},
				Labels: defaultCoinStatsLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "rank"),
					"Current coin position by market cap",
					defaultCoinStatsLabels, nil,
				),
				Value: func(coinStat Coin) float64 {
					return float64(coinStat.Rank)
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

	// Instantiate default collector
	co := colly.NewCollector()

	co.OnHTML("tbody tr", func(e *colly.HTMLElement) {
		if e.ChildText(".cmc-table__column-name") != "" {
			coin := parseTable(e)
			coins = append(coins, coin)
		}
	})

	co.Visit(c.url)

	return coins, nil
}

func toFloat(input string) float64 {
	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	return value
}

func parsePrice(input string) float64 {
	replacer := strings.NewReplacer("$", "", ",", "")
	dollarValue := replacer.Replace(input)
	return toFloat(dollarValue)
}

func parseMarketCap(input string) float64 {
	reg, err := regexp.Compile("^\\$.*\\$")
	if err != nil {
		fmt.Println(err)
		return -1
	}
	marketCapValue := reg.ReplaceAllString(input, "")

	replacer := strings.NewReplacer("$", "", ",", "")
	parsedValue := replacer.Replace(marketCapValue)
	return toFloat(parsedValue)
}

func parseSupply(input string) float64 {
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		fmt.Println(err)
		return -1
	}
	supplyValue := reg.ReplaceAllString(input, "")
	return toFloat(supplyValue)
}

func parsePercentage(input string) float64 {
	reg, err := regexp.Compile("[^0-9\\.\\-]")
	if err != nil {
		fmt.Println(err)
	}
	percentValue := reg.ReplaceAllString(input, "")
	return toFloat(percentValue)
}

func parseTable(e *colly.HTMLElement) Coin {
	var coin Coin

	coin.Rank, _ = strconv.ParseFloat(e.ChildText(".cmc-table__cell--sort-by__rank"), 64)
	coin.Symbol = e.ChildText(".cmc-table__cell--sort-by__symbol")
	coin.Market_Cap = parseMarketCap(e.ChildText(".cmc-table__cell--sort-by__market-cap"))
	coin.Price = parsePrice(e.ChildText(".cmc-table__cell--sort-by__price"))
	coin.Circulating_Supply = parseSupply(e.ChildText(".cmc-table__cell--sort-by__circulating-supply"))
	coin.Volume_24h = parsePrice(e.ChildText(".cmc-table__cell--sort-by__volume-24-h"))
	coin.Change_1h = parsePercentage(e.ChildText(".cmc-table__cell--sort-by__percent-change-1-h"))
	coin.Change_24h = parsePercentage(e.ChildText(".cmc-table__cell--sort-by__percent-change-24-h"))
	coin.Change_7d = parsePercentage(e.ChildText(".cmc-table__cell--sort-by__percent-change-7-d"))

	return coin
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
