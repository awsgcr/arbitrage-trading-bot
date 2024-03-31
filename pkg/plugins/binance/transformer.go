package binance

import (
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
)

func convertToUserDataEvent(bEvent *BWsUserDataEvent) *UserDataEvent {
	event := &UserDataEvent{
		Event:             UserDataEventType(bEvent.Event),
		Time:              uint64(bEvent.Time),
		TransactionTime:   bEvent.TransactionTime,
		AccountUpdateTime: bEvent.AccountUpdateTime,
		AccountUpdate:     convertToAccountUpdate(bEvent.AccountUpdate),
		BalanceUpdate:     convertToBalanceUpdate(bEvent.BalanceUpdate),
		OrderUpdate:       convertToOrderUpdate(bEvent.OrderUpdate),
		OCOUpdate:         convertToOCOUpdate(bEvent.OCOUpdate),
	}
	return event
}

func convertToOCOUpdate(update BWsOCOUpdate) WsOCOUpdate {
	return WsOCOUpdate{
		Symbol:          update.Symbol,
		OrderListId:     update.OrderListId,
		ContingencyType: update.ContingencyType,
		ListStatusType:  update.ListStatusType,
		ListOrderStatus: update.ListOrderStatus,
		RejectReason:    update.RejectReason,
		ClientOrderId:   update.ClientOrderId,
		Orders:          convertToWsOCOOrderList(update.Orders),
	}
}

func convertToWsOCOOrderList(orders BWsOCOOrderList) WsOCOOrderList {
	size := len(orders.WsOCOOrders)
	var wsOCOOrders = make([]WsOCOOrder, size)
	for i := 0; i < size; i++ {
		item := orders.WsOCOOrders[i]
		wsOCOOrders[i] = WsOCOOrder{
			Symbol:        item.Symbol,
			OrderId:       item.OrderId,
			ClientOrderId: item.ClientOrderId,
		}
	}
	return WsOCOOrderList{
		WsOCOOrders: wsOCOOrders,
	}
}

func convertToOrderUpdate(update BWsOrderUpdate) WsOrderUpdate {
	//b, _ := json.Marshal(update)
	//lg.Info("print orderUpdate of binance", "orderUpdate", string(b))
	return WsOrderUpdate{
		Symbol:        newSymbolFromString(update.Symbol),
		ClientOrderId: update.ClientOrderId,
		Side:          SideType(update.Side),
		Type:          OrderType(update.Type),
		TimeInForce:   TimeInForceType(update.TimeInForce),
		Volume:        NewDecimalFromStringIgnoreErr(update.Volume),
		Price:         NewDecimalFromStringIgnoreErr(update.Price),
		LatestPrice:   NewDecimalFromStringIgnoreErr(update.LatestPrice),
		//StopPrice:               NewDecimalFromStringIgnoreErr(update.StopPrice),
		//TrailingDelta:           update.TrailingDelta,
		//IceBergVolume:           update.IceBergVolume,
		//OrderListId:             update.OrderListId,
		//OrigCustomOrderId:       update.OrigCustomOrderId,
		//ExecutionType:           update.ExecutionType,
		Status:       OrderStatusType(update.Status),
		RejectReason: update.RejectReason,
		Id:           update.Id,
		//LatestVolume:            update.LatestVolume,
		FilledVolume: NewDecimalFromStringIgnoreErr(update.FilledVolume),
		//LatestPrice:             NewDecimalFromStringIgnoreErr(update.LatestPrice),
		//FeeAsset:                update.FeeAsset,
		//FeeCost:                 NewDecimalFromStringIgnoreErr(update.FeeCost),
		TransactionTime: update.TransactionTime,
		//TradeId:                 update.TradeId,
		//IsInOrderBook:           update.IsInOrderBook,
		IsMaker:           update.IsMaker,
		CreateTime:        update.CreateTime,
		FilledQuoteVolume: NewDecimalFromStringIgnoreErr(update.FilledQuoteVolume),
		//LatestQuoteVolume:       update.LatestQuoteVolume,
		//QuoteVolume:             update.QuoteVolume,
		//TrailingTime:            update.TrailingTime,
		//StrategyId:              update.StrategyId,
		//StrategyType:            update.StrategyType,
		//WorkingTime:             update.WorkingTime,
		//SelfTradePreventionMode: update.SelfTradePreventionMode,
	}
}

func convertToBalanceUpdate(update BWsBalanceUpdate) WsBalanceUpdate {
	return WsBalanceUpdate{
		Asset:  ToAsset(update.Asset),
		Change: NewDecimalFromStringIgnoreErr(update.Change),
	}
}

func convertToAccountUpdate(update BWsAccountUpdateList) WsAccountUpdateList {
	size := len(update.WsAccountUpdates)
	var wsAccountUpdate = make([]WsAccountUpdate, size)
	for i := 0; i < size; i++ {
		item := update.WsAccountUpdates[i]
		wsAccountUpdate[i] = WsAccountUpdate{
			Asset:  ToAsset(item.Asset),
			Free:   NewDecimalFromStringIgnoreErr(item.Free),
			Locked: NewDecimalFromStringIgnoreErr(item.Locked),
		}
	}
	return WsAccountUpdateList{
		WsAccountUpdates: wsAccountUpdate,
	}
}
