package moon

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/toorop/go-bittrex"
)

type Bittrex struct {
}

// Bittrex pair format is BTC-XRP
func (b *Bittrex) nativePair(pair string) string {
	pairs := strings.SplitN(pair, "_", 2)
	return pairs[1] + "-" + pairs[0]
}

func (b *Bittrex) GetTicker(ctx context.Context, pair string) (*Ticker, error) {
	bittrex := bittrex.New("", "")
	mss, err := bittrex.GetMarketSummary(b.nativePair(pair))
	if err != nil {
		return nil, err
	}
	if len(mss) < 1 {
		return nil, errors.New("doesn't response from Bittrex")
	}
	ms := mss[0]

	return &Ticker{
		Bid:       FormatFloat(ms.Bid),
		Ask:       FormatFloat(ms.Ask),
		Last:      FormatFloat(ms.Last),
		Volume:    FormatFloat(ms.Volume),
		Timestamp: b.parseTimestamp(ms.TimeStamp),
	}, nil
}

func (b *Bittrex) parseTimestamp(ts string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.99", ts)
	return t
}

func (b *Bittrex) GetBalances(ctx context.Context, key string, secret string) (*Balances, error) {
	bittrex := bittrex.New(key, secret)
	balances, err := bittrex.GetBalances()
	if err != nil {
		return nil, err
	}

	ret := make(Balances)
	for _, balance := range balances {
		if balance.Balance == 0 {
			continue
		}

		symbol := balance.Currency // XRP
		btcValue, err := b.calcBtcValue(bittrex, symbol, balance.Balance)
		if err != nil {
			return nil, err
		}

		ret[symbol] = Balance{
			Available: FormatFloat(balance.Available),
			OnOrders:  FormatFloat(balance.Balance - balance.Available),
			Amount:    FormatFloat(balance.Balance),
			BtcValue:  FormatFloat(btcValue),
		}
	}
	return &ret, nil
}

func (b *Bittrex) calcBtcValue(bittrex *bittrex.Bittrex, symbol string, amount float64) (float64, error) {
	if symbol == "BTC" || symbol == "btc" {
		return amount, nil
	}
	pair := "BTC-" + symbol
	inverse := false

	if pair == "BTC-USDT" {
		pair = "USDT-BTC"
		inverse = true
	}

	ticker, err := bittrex.GetTicker(pair)
	if err != nil {
		return 0, err
	}

	rate := ticker.Last
	if inverse {
		rate = 1 / rate
	}
	return amount * rate, nil
}
