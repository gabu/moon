package moon

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gabu/go-liqui"
)

type Liqui struct {
}

// Liqui pair format is eth_btc
func (l *Liqui) nativePair(pair string) string {
	return strings.ToLower(pair)
}

func (l *Liqui) GetTicker(ctx context.Context, pair string) (*Ticker, error) {
	pair = l.nativePair(pair)
	liqui := liqui.NewClient()
	tickers, err := liqui.Ticker(ctx, pair)
	if err != nil {
		return nil, err
	}

	ticker, ok := tickers[pair]
	if !ok {
		return nil, errors.New("Error: " + pair + " is not found")
	}

	timestamp := time.Unix(int64(ticker.Updated), 0)

	return &Ticker{
		Bid:       FormatFloat(ticker.Buy),
		Ask:       FormatFloat(ticker.Sell),
		Last:      FormatFloat(ticker.Last),
		Volume:    FormatFloat(ticker.Vol),
		Timestamp: timestamp,
	}, nil
}

func (l *Liqui) GetBalances(ctx context.Context, key string, secret string) (*Balances, error) {
	liqui := liqui.NewClient().Auth(key, secret)
	info, err := liqui.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	// symbol with total of on orders amount
	onOrdersList := map[string]float64{}
	if info.OpenOrders > 0 {
		orders, err := liqui.ActiveOrders(ctx, "")
		if err != nil {
			return nil, err
		}
		for _, order := range orders {
			symbols := strings.SplitN(order.Pair, "_", 2)
			symbol := symbols[0] // sell
			if order.Type == "buy" {
				symbol = symbols[1]
			}
			if _, ok := onOrdersList[symbol]; ok {
				onOrdersList[symbol] += order.Amount
			} else {
				onOrdersList[symbol] = order.Amount
			}
		}
	}

	amounts := map[string]float64{}
	for symbol, available := range info.Funds {
		onOrders := onOrdersList[symbol]
		amount := available + onOrders
		if amount == 0 {
			continue
		}
		amounts[symbol] = amount
	}

	btcValues, err := l.calcBtcValues(ctx, liqui, amounts)
	if err != nil {
		return nil, err
	}

	ret := make(Balances)
	for symbol, amount := range amounts {
		available := info.Funds[symbol]
		onOrders := onOrdersList[symbol]
		btcValue := btcValues[symbol]

		ret[strings.ToUpper(symbol)] = Balance{
			Available: FormatFloat(available),
			OnOrders:  FormatFloat(onOrders),
			Amount:    FormatFloat(amount),
			BtcValue:  FormatFloat(btcValue),
		}
	}
	return &ret, nil
}

func (l *Liqui) calcBtcValues(ctx context.Context, liqui *liqui.Client, amounts map[string]float64) (map[string]float64, error) {
	ret := map[string]float64{}
	pairs := ""
	// create pairs string
	for symbol := range amounts {
		if symbol == "btc" {
			continue
		}
		if pairs == "" {
			pairs = symbol + "_btc"
		} else {
			pairs += "-" + symbol + "_btc"
		}
	}

	tickers, err := liqui.Ticker(ctx, pairs)
	if err != nil {
		return ret, err
	}

	for symbol, amount := range amounts {
		if symbol == "btc" {
			ret[symbol] = amount
			continue
		}

		pair := symbol + "_btc"
		ticker, ok := tickers[pair]
		if !ok {
			return ret, errors.New(pair + " is not found")
		}
		ret[symbol] = amount * ticker.Last
	}
	return ret, nil
}
