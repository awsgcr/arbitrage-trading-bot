package alerting

import (
	"fmt"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/components/simplejson"
	"jasonzhu.com/coin_labor/core/setting"
	"jasonzhu.com/coin_labor/core/util/http"
	"strings"
	"sync"
	"time"
)

const (
	tgDomain = "https://api.telegram.org"
)

var defaultTelegramNotifier *TelegramNotifier

func init() {
	defaultTelegramNotifier = &TelegramNotifier{
		lg:      log.New("alerting.notifier.tg"),
		msgChan: make(chan *Msg),
	}
	defaultTelegramNotifier.Run()
}

type TelegramNotifier struct {
	lg    log.Logger
	msgMu sync.Mutex

	msgChan  chan *Msg
	msgCache []*Msg
}

func (n *TelegramNotifier) InfoNotify(msg string, params ...interface{}) (err error) {
	return n.notify(infoType, msg, params...)
}

func (n *TelegramNotifier) ErrorNotify(msg string, params ...interface{}) (err error) {
	return n.notify(errorType, msg, params...)
}

func (n *TelegramNotifier) InfoNotifyRightNow(msg string, params ...interface{}) (err error) {
	err = n.InfoNotify(msg, params...)
	n.Flush()
	return err
}

func (n *TelegramNotifier) ErrorNotifyRightNow(msg string, params ...interface{}) (err error) {
	err = n.ErrorNotify(msg, params...)
	n.Flush()
	return err
}

func (n *TelegramNotifier) notify(typ msgType, msg string, params ...interface{}) (err error) {
	if !setting.AlertingEnabled {
		n.lg.Info("TelegramNotifier is off.")
		return nil
	}

	message, err := buildContent(msg, params...)
	if err != nil {
		n.lg.Error("buildContent error",
			"msg", msg,
			"params", params,
		)
		n.msgChan <- &Msg{
			typ:  typ,
			data: fmt.Sprintf("buildContent error, msg: %s, type: %d", msg, typ),
		}
		return err
	}

	n.msgChan <- &Msg{
		typ:  typ,
		data: message,
	}
	return nil
}

func (n *TelegramNotifier) Flush() {
	if !setting.AlertingEnabled {
		n.lg.Info("TelegramNotifier is off.")
		return
	}

	n.sendAllInCache()
}

func (n *TelegramNotifier) Run() {
	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for {
			select {
			case msg := <-n.msgChan:
				n.addMsg2Cache(msg)
			case <-ticker.C:
				n.sendAllInCache()
			}
		}
	}()
}
func (n *TelegramNotifier) addMsg2Cache(msg *Msg) {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()

	n.msgCache = append(n.msgCache, msg)
}

func (n *TelegramNotifier) sendAllInCache() {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()

	var infoArr, warnArr, errorArr []string
	for _, msg := range n.msgCache {
		switch msg.typ {
		case infoType:
			infoArr = append(infoArr, msg.data)
		case warnType:
			warnArr = append(warnArr, msg.data)
		case errorType:
			errorArr = append(errorArr, msg.data)
		}
	}

	_ = n.sendMessages(infoType, infoArr)
	_ = n.sendMessages(warnType, warnArr)
	_ = n.sendMessages(errorType, errorArr)
	n.msgCache = n.msgCache[0:0]
	return
}

func (n *TelegramNotifier) sendMessages(typ msgType, msgArr []string) error {
	if len(msgArr) == 0 {
		return nil
	}
	msgArr = append(msgArr, "")
	msg := strings.Join(msgArr, "\\n--------------------------------\\n")
	err := n.sendMsg(typ, msg)
	if err != nil {
		n.lg.Error("Send TG Error", "err", err)
		errMsg, _ := buildContent("Send TG Error", "err", err)
		_ = n.sendMsg(warnType, errMsg)
	}
	return err
}

func (n *TelegramNotifier) sendMsg(typ msgType, msg string) error {
	bodyStr := `{"chat_id": "-828188094", "text": "` + msg + `", "disable_notification": true}`
	bodyJSON, err := simplejson.NewJson([]byte(bodyStr))
	if err != nil {
		n.lg.Error("Failed to create Json data", "error", err)
		return err
	}
	body, err := bodyJSON.MarshalJSON()
	if err != nil {
		return err
	}

	url := buildNotifierURL(setting.TgToken)
	_, err = http.PostJson(url, string(body))
	if err != nil {
		n.lg.Error("Failed to send TG", "error", err)
		return err
	}

	return nil
}

func buildNotifierURL(tgToken string) string {
	var tgNotifierURL = fmt.Sprintf("%s/bot%s/sendMessage", tgDomain, tgToken)
	return tgNotifierURL
}
