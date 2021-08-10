package collector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/log"
)

var coins = []Coin{
	{
		Symbol:             "ADA",
		Market_Cap:         47185315258,
		Price:              1.47,
		Circulating_Supply: 32112395093,
		Volume_24h:         1782521919,
		Change_1h:          0.11,
		Change_24h:         4.3,
		Change_7d:          13.17,
	},
	{
		Symbol:             "BTC",
		Market_Cap:         852656824242,
		Price:              45398.34,
		Circulating_Supply: 18781675,
		Volume_24h:         38394061287,
		Change_1h:          -0.16,
		Change_24h:         4.62,
		Change_7d:          17.76,
	},
}

func TestCoinStats(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, webResponse)
	}))
	defer ts.Close()

	_, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %s", err)
	}

	c := NewCoinStats(log.NewNopLogger(), http.DefaultClient, ts.URL)
	coinStatsResp, err := c.fetchAndDecodeCoinStats()
	if err != nil {
		t.Fatalf("Failed to fetch or decode coin stats: %s", err)
	}

	for _, coin := range coins {
		for _, stats := range coinStatsResp {
			if coin.Symbol == stats.Symbol {
				if coin.Price != stats.Price {
					t.Errorf("Wrong Price for symbol '%s'. Expected: %f, Received: %f", coin.Symbol, coin.Price, stats.Price)
				}
				if coin.Market_Cap != stats.Market_Cap {
					t.Errorf("Wrong Market Cap for symbol '%s'. Expected: %f, Received: %f", coin.Symbol, coin.Market_Cap, stats.Market_Cap)
				}
				if coin.Circulating_Supply != stats.Circulating_Supply {
					t.Errorf("Wrong Circulating Supply for symbol '%s'. Expected: %f, Received: %f", coin.Symbol, coin.Circulating_Supply, stats.Circulating_Supply)
				}
				if coin.Volume_24h != stats.Volume_24h {
					t.Errorf("Wrong Volume for symbol '%s'. Expected: %f, Received: %f", coin.Symbol, coin.Volume_24h, stats.Volume_24h)
				}
				if coin.Change_1h != stats.Change_1h {
					t.Errorf("Wrong Change 1h for symbol '%s'. Expected: %f, Received: %f", coin.Symbol, coin.Change_1h, stats.Change_1h)
				}
				if coin.Change_24h != stats.Change_24h {
					t.Errorf("Wrong Change 24h for symbol '%s'. Expected: %f, Received: %f", coin.Symbol, coin.Change_24h, stats.Change_24h)
				}
				if coin.Change_7d != stats.Change_7d {
					t.Errorf("Wrong Change 7d for symbol '%s'. Expected: %f, Received: %f", coin.Symbol, coin.Change_7d, stats.Change_7d)
				}
			}
		}
	}
}
