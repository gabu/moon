package moon

import (
	"context"
	"strconv"
	"time"
)

type Exchange interface {
	GetTicker(ctx context.Context, pair string) (*Ticker, error)
	GetBalances(ctx context.Context, key string, secret string) (*Balances, error)
}

type Ticker struct {
	Bid       string
	Ask       string
	Last      string
	Volume    string
	Timestamp time.Time
}

type Balances map[string]Balance

type Balance struct {
	Available string
	OnOrders  string
	Amount    string
	BtcValue  string
}

func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 8, 64)
}
