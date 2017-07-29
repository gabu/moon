package moon

import (
	"context"
	"strconv"
	"strings"

	"github.com/bitfinexcom/bitfinex-api-go/v1"
)

type Bitfinex struct {
}

// Bitfinex pair format is xrpbtc
func (b *Bitfinex) nativePair(pair string) string {
	return strings.ToLower(strings.Replace(pair, "_", "", 1))
}

func (b *Bitfinex) GetTicker(ctx context.Context, pair string) (*Ticker, error) {
	bitfinex := bitfinex.NewClient()
	ticker, err := bitfinex.Ticker.Get(b.nativePair(pair))
	if err != nil {
		return nil, err
	}

	timestamp, err := ticker.ParseTime()
	if err != nil {
		return nil, err
	}

	return &Ticker{
		Bid:       ticker.Bid,
		Ask:       ticker.Ask,
		Last:      ticker.LastPrice,
		Volume:    ticker.Volume,
		Timestamp: *timestamp,
	}, nil
}

func (b *Bitfinex) GetBalances(ctx context.Context, key string, secret string) (*Balances, error) {
	bitfinex := bitfinex.NewClient().Auth(key, secret)
	balances, err := bitfinex.Balances.All()
	if err != nil {
		return nil, err
	}

	ret := make(Balances)
	for _, balance := range balances {
		// Type is including three wallets
		// "trading" is Margin Wallet
		// "deposit" is Funding Wallet
		// "exchange" is Exchange Wallet
		// currentry supports only Exchange Wallet now.
		if balance.Type == "trading" || balance.Type == "deposit" {
			continue
		}

		symbol := strings.ToUpper(balance.Currency)
		amount, err := strconv.ParseFloat(balance.Amount, 64)
		if err != nil {
			return nil, err
		}
		available, err := strconv.ParseFloat(balance.Available, 64)
		if err != nil {
			return nil, err
		}
		btcValue, err := b.calcBtcValue(bitfinex, symbol, amount)
		if err != nil {
			return nil, err
		}

		ret[symbol] = Balance{
			Available: balance.Available,
			OnOrders:  FormatFloat(amount - available),
			Amount:    balance.Amount,
			BtcValue:  FormatFloat(btcValue),
		}
	}
	return &ret, nil
}

func (b *Bitfinex) calcBtcValue(bitfinex *bitfinex.Client, symbol string, amount float64) (float64, error) {
	if symbol == "BTC" || symbol == "btc" {
		return amount, nil
	}
	pair := strings.ToLower(symbol) + "btc"
	ticker, err := bitfinex.Ticker.Get(pair)
	if err != nil {
		return 0, err
	}
	last, err := strconv.ParseFloat(ticker.LastPrice, 64)
	if err != nil {
		return 0, err
	}
	return amount * last, nil
}
