package general

import (
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
)

// Account define account info
type Account struct {
	MakerCommission  int64  `json:"makerCommission"`
	TakerCommission  int64  `json:"takerCommission"`
	BuyerCommission  int64  `json:"buyerCommission"`
	SellerCommission int64  `json:"sellerCommission"`
	CanTrade         bool   `json:"canTrade"`
	CanWithdraw      bool   `json:"canWithdraw"`
	CanDeposit       bool   `json:"canDeposit"`
	UpdateTime       uint64 `json:"updateTime"`
	AccountType      string `json:"accountType"`

	BalancesMap map[Asset]Balance
	rwM         sync.RWMutex
}

// Balance define user balance of your account
type Balance struct {
	Asset  Asset
	Free   decimal.Decimal
	Locked decimal.Decimal
}

// NewAccount Create Account /*********************************************/
func NewAccount(time uint64, balances []Balance) *Account {
	a := &Account{
		UpdateTime: time,
	}
	a.InitBalances(balances)
	return a
}

func (a *Account) InitBalances(balances []Balance) {
	a.rwM.Lock()
	a.BalancesMap = make(map[Asset]Balance)
	for _, balance := range balances {
		a.BalancesMap[balance.Asset] = balance
	}
	a.rwM.Unlock()
}

func (a *Account) GetBalance(asset Asset) *Balance {
	a.rwM.RLock()
	defer func() {
		a.rwM.RUnlock()
	}()

	if a.BalancesMap == nil {
		return NewBalanceWithZero(asset)
	}
	if balance, ok := a.BalancesMap[asset]; ok {
		return &balance
	}
	return NewBalanceWithZero(asset)
}

func (a *Account) UpdateBalances(exchange Exchange, time uint64, update WsAccountUpdateList) {
	if time < a.UpdateTime {
		return
	}

	a.rwM.Lock()
	defer func() {
		a.rwM.Unlock()
	}()

	for _, au := range update.WsAccountUpdates {
		balance := Balance{
			Asset:  au.Asset,
			Free:   au.Free,
			Locked: au.Locked,
		}
		a.BalancesMap[au.Asset] = balance
		glg.Warn("balance updated", "exchange", exchange, "newBalance", balance.ToString())
	}
	a.UpdateTime = time
}

// NewBalanceWithZero /*********************************************/
func NewBalanceWithZero(asset Asset) *Balance {
	return &Balance{
		Asset:  asset,
		Free:   decimal.Zero,
		Locked: decimal.Zero,
	}
}

func (b *Balance) HasFree() bool {
	return b.Free.GreaterThan(decimal.Zero)
}
func (b *Balance) HasLocked() bool {
	return b.Locked.GreaterThan(decimal.Zero)
}

func (b *Balance) Total() decimal.Decimal {
	return b.Free.Add(b.Locked)
}

func (b *Balance) ToString() string {
	return fmt.Sprintf("asset: %s, free: %s, locked: %s, total: %s", b.Asset, b.Free, b.Locked, b.Total())
}

// IsSupportedAsset /*********************************************/
func IsSupportedAsset(s string, supportedAssets []Asset) (Asset, bool) {
	for _, asset := range supportedAssets {
		if strings.ToUpper(string(asset)) == strings.ToUpper(s) {
			return asset, true
		}
	}
	return UnKnown, false
}

func ToAsset(s string) Asset {
	return Asset(strings.ToUpper(s))
}
