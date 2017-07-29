package moon

import (
	"context"
	"strconv"
	"strings"

	"github.com/gabu/go-poloniex"
)

type Poloniex struct {
}

// Poloniex pair format is BTC_XRP
func (p *Poloniex) nativePair(pair string) string {
	pairs := strings.SplitN(pair, "_", 2)
	return pairs[1] + "_" + pairs[0]
}

func (p *Poloniex) GetTicker(ctx context.Context, pair string) (*Ticker, error) {
	poloniex, err := poloniex.New("", "", "", nil)
	if err != nil {
		return nil, err
	}

	tickers, err := poloniex.GetTickers(ctx)
	if err != nil {
		return nil, err
	}

	ticker := tickers.Pair[p.nativePair(pair)]

	return &Ticker{
		Bid:       ticker.HighestBid,
		Ask:       ticker.LowestAsk,
		Last:      ticker.Last,
		Volume:    ticker.BaseVolume,
		Timestamp: ticker.Time,
	}, nil
}

func (p *Poloniex) GetBalances(ctx context.Context, key string, secret string) (*Balances, error) {
	poloniex, err := poloniex.New(key, secret, "", nil)
	if err != nil {
		return nil, err
	}

	balances, err := poloniex.GetCompleteBalances(ctx)
	if err != nil {
		return nil, err
	}

	ret := make(Balances)
	for k, b := range balances {
		available, err := strconv.ParseFloat(b.Available, 64)
		if err != nil {
			return nil, err
		}
		onOrders, err := strconv.ParseFloat(b.OnOrders, 64)
		if err != nil {
			return nil, err
		}
		amount := available + onOrders
		if amount == 0 {
			continue
		}

		ret[k] = Balance{
			Available: b.Available,
			OnOrders:  b.OnOrders,
			Amount:    FormatFloat(amount),
			BtcValue:  b.BtcValue,
		}
	}
	return &ret, nil
}
