package moon

import (
	"context"
	"strings"
	"time"

	"github.com/gabu/go-bitgrail"
)

type Bitgrail struct {
}

// BitGrail pair format is BTC-XRB
func (b *Bitgrail) nativePair(pair string) string {
	pairs := strings.SplitN(pair, "_", 2)
	return pairs[1] + "-" + pairs[0]
}

func (b *Bitgrail) GetTicker(ctx context.Context, pair string) (*Ticker, error) {
	client := bitgrail.NewClient()
	ticker, err := client.Ticker(ctx, b.nativePair(pair))
	if err != nil {
		return nil, err
	}

	return &Ticker{
		Bid:       FormatFloat(ticker.Bid),
		Ask:       FormatFloat(ticker.Ask),
		Last:      FormatFloat(ticker.Last),
		Volume:    FormatFloat(ticker.Volume),
		Timestamp: time.Now(),
	}, nil
}

func (b *Bitgrail) GetBalances(ctx context.Context, key string, secret string) (*Balances, error) {
	client := bitgrail.NewClient().Auth(key, secret)
	balances, err := client.Balances(ctx)
	if err != nil {
		return nil, err
	}

	ret := make(Balances)
	for symbol, balance := range balances {
		amount := balance.Balance + balance.Reserved
		if amount == 0 {
			continue
		}

		btcValue, err := b.calcBtcValue(ctx, client, symbol, amount)
		if err != nil {
			return nil, err
		}

		ret[symbol] = Balance{
			Available: FormatFloat(balance.Balance),
			OnOrders:  FormatFloat(balance.Reserved),
			Amount:    FormatFloat(amount),
			BtcValue:  FormatFloat(btcValue),
		}
	}
	return &ret, nil
}

func (b *Bitgrail) calcBtcValue(ctx context.Context, client *bitgrail.Client, symbol string, amount float64) (float64, error) {
	if symbol == "BTC" || symbol == "btc" {
		return amount, nil
	}
	pair := "BTC-" + symbol
	ticker, err := client.Ticker(ctx, pair)
	if err != nil {
		return 0, err
	}
	return amount * ticker.Last, nil
}
