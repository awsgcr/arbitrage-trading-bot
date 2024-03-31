package general

import (
	"context"
	"reflect"
	"sort"
)

var plugins []*ExPlugin
var pluginsMap = make(map[Exchange]ExManager)

type ExPlugin struct {
	ExName   Exchange
	Instance ExManager
	Ranking  int
}

func RegisterPlugin(instance ExManager) {
	plugin := ExPlugin{
		ExName:   Exchange(reflect.TypeOf(instance).Elem().Name()),
		Instance: instance,
		Ranking:  500,
	}
	plugins = append(plugins, &plugin)
	pluginsMap[plugin.ExName] = instance
}

func Register(plugin *ExPlugin) {
	plugins = append(plugins, plugin)
	pluginsMap[plugin.ExName] = plugin.Instance
}

func GetExPlugins() []*ExPlugin {
	slice := plugins
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Ranking > slice[j].Ranking
	})

	return slice
}

func GetExPluginByExchange(name Exchange) ExManager {
	return pluginsMap[name]
}

type ExManager interface {
	ExchangeAlias() Exchange
	GetBaseInfoManager() BaseInterface
	GetMarketInfoManager() MarketInterface
	GetAccountManager() AccountInterface
	GetOrderInterface() OrderInterface
}

type BaseInterface interface {
	ServerTime() (int64, error)
	GetSymbolBasicInfo(symbol Symbol) (*SymbolBasicInfo, error)
	GetSymbolsBasicInfo() map[Symbol]*SymbolBasicInfo
}

type MarketInterface interface {
	FetchDepth(symbol Symbol, limit int) (*DepthInfo, error)
	WsWatchMarketDepth(ctx context.Context, infoC chan *DepthInfo, symbols ...Symbol) error
}

type AccountInterface interface {
	GetAccountInfo() (*Account, error)
	GetBalanceAtStart(symbol Asset) *Balance
	WsWatchUserDataChanges(ctx context.Context, eventC chan *UserDataEvent) error
}

type OrderInterface interface {
	ListOpenOrdersOfSymbol(symbol Symbol) (res []*Order, err error)
	ListAllOrders(symbol Symbol) ([]*Order, error)
	CreateOrder(plan OrderPlan) (*CreateOrderResponse, error)
	GetOrder(symbol Symbol, orderId string, clientOrderId string) (*Order, error)
	CancelOrder(symbol Symbol, orderId string, clientOrderId string) (OrderStatusType, error)
}

type MarketWatchInterface interface {
	WatchDepth(ctx context.Context, symbols ...Symbol) error
}
