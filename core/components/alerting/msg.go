package alerting

const (
	infoType  msgType = 1
	warnType  msgType = 2
	errorType msgType = 3
)

type msgType int

type Msg struct {
	typ  msgType
	data string
}
