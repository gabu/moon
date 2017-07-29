package moon

import (
	"context"
	"strings"
	"time"

	"github.com/gabu/go-cryptopia"
)

type Cryptopia struct {
}

func (c *Cryptopia) GetTicker(ctx context.Context, pair string) (*Ticker, error) {
	client := cryptopia.NewClient()
	market, err := client.GetMarket(ctx, pair, 0)
	if err != nil {
		return nil, err
	}

	return &Ticker{
		Bid:       FormatFloat(market.BidPrice),
		Ask:       FormatFloat(market.AskPrice),
		Last:      FormatFloat(market.LastPrice),
		Volume:    FormatFloat(market.Volume),
		Timestamp: time.Now(),
	}, nil
}

func (c *Cryptopia) GetBalances(ctx context.Context, key string, secret string) (*Balances, error) {
	client := cryptopia.NewClient().Auth(key, secret)
	balances, err := client.GetBalance(ctx)
	if err != nil {
		return nil, err
	}

	amounts := map[string]float64{}
	ret := Balances{}
	for _, b := range balances {
		if b.Total == 0 {
			continue
		}
		amounts[b.Symbol] = b.Total
		ret[b.Symbol] = Balance{
			Available: FormatFloat(b.Available),
			OnOrders:  FormatFloat(b.HeldForTrades),
			Amount:    FormatFloat(b.Total),
		}
	}

	btcValues, err := c.calcBtcValues(ctx, client, amounts)
	if err != nil {
		return nil, err
	}

	for symbol, btcValue := range btcValues {
		balance := ret[symbol]
		balance.BtcValue = FormatFloat(btcValue)
		ret[symbol] = balance
	}
	return &ret, nil
}

func (c *Cryptopia) calcBtcValues(ctx context.Context, client *cryptopia.Client, amounts map[string]float64) (map[string]float64, error) {
	ret := map[string]float64{}
	markets, err := client.GetMarkets(ctx, "BTC", 0)
	if err != nil {
		return ret, err
	}

	if amount, ok := amounts["BTC"]; ok {
		ret["BTC"] = amount
	}

	for _, market := range markets {
		symbol := strings.SplitN(market.Label, "/", 2)[0]
		if amount, ok := amounts[symbol]; ok {
			ret[symbol] = amount * market.LastPrice
		}
	}
	return ret, nil
}
