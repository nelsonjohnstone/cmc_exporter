# CMC Exporter

Prometheus exporter for various cryptocurrency metrics, written in Go.

#### Docker

```bash
docker build -t cmc-exporter:latest
docker run --rm -p 9599:9599 nelsonjohnstone/cmc-exporter:latest
```

Example `docker-compose.yml`:

```yaml
elasticsearch_exporter:
    image: nelsonjohnstone/cmc-exporter:latest
    restart: always
    ports:
    - "9599:9599"
```

### Configuration

Below is the command line options summary:
```bash
cmc_exporter --help
```

| Argument                | Description | Default     |
| --------                | ----------- | ----------- |
| cmc.uri                 | Uri of the website to collect crypto data from. | https://coinmarketcap.com/all/views/all/ |
| cmc.timeout             | Timeout for trying to get stats from the website. (eg: 20s) | 5s |
| web.listen-address      | Address to listen on for web interface and telemetry. | :9599 |
| web.telemetry-path      | Path under which to expose metrics. | /metrics |
| version                 | Show version info on stdout and exit. | |

Commandline parameters are specified with `--`. Also, all commandline parameters can be provided as environment variables. The environment variable name is derived from the parameter name
by replacing `.` and `-` with `_` and upper-casing the parameter name.

### Metrics

|Name                                                                   |Type       |Help
|----                                                                   |----       |----
| cmc_coin_stats_change_1h                                              | gauge     | Percentage change in price over the last 1 hour
| cmc_coin_stats_change_24h                                             | gauge     | Percentage change in price over the last 24 hours
| cmc_coin_stats_change_7d                                              | gauge     | Percentage change in price over the last 7 days
| cmc_coin_stats_circulating_supply                                     | gauge     | The amount of coins that are circulating in the market and are in public hands
| cmc_coin_stats_market_cap                                             | gauge     | The total market value of a cryptocurrency's circulating supply
| cmc_coin_stats_price                                                  | gauge     | Current price in USD
| cmc_coin_stats_rank                                                   | gauge     | Current coin position by market cap
| cmc_coin_stats_volume_24hr                                            | gauge     | A measure of how much of a cryptocurrency was traded in the last 24 hours
| cmc_coin_stats_total_scrapes                                          | counter   | Current total website scrapes
| cmc_coin_stats_html_parse_failures                                    | gauge     | Number of errors while parsing HTML
| cmc_coin_stats_up                                                     | gauge     | Was the last scrape of the website successful
| cmc_exporter_build_info                                               | gauge     | A metric with a constant '1' value labeled by version, revision, branch, and goversion from which cmc_exporter was built
