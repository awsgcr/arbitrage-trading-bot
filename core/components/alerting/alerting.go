package alerting

import "jasonzhu.com/coin_labor/core/components/log"

/**

@author Jason
@version 2020-07-23 22:25
*/

var lg = log.New("alerting")

func Notify(err error, msg string, params ...interface{}) {
	go func() {
		notify(err, msg, params...)
	}()
}

func Info(msg string, params ...interface{}) {
	go func() {
		notify(nil, msg, params...)
	}()
}

func NotifyRightNow(err error, msg string, params ...interface{}) {
	go func() {
		notify(err, msg, params...)
		flush()
	}()
}

func Flush() {
	go func() {
		flush()
	}()
}

func notify(err error, msg string, params ...interface{}) {
	if err != nil {
		params = append(params, "err", err)
		lg.Error(msg, params...)
		_ = defaultTelegramNotifier.ErrorNotify(msg, params...)
	} else {
		lg.Info(msg, params...)
		_ = defaultTelegramNotifier.InfoNotify(msg, params...)
	}
}

func flush() {
	defaultTelegramNotifier.Flush()
}
