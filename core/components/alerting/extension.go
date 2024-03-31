package alerting

import (
	"fmt"
	"strings"
	"time"
)

var START_TIME = time.Now()

func buildContent(msg string, params ...interface{}) (string, error) {
	if len(params)%2 != 0 {
		return "", fmt.Errorf("length of param could not be odd number")
	}
	slice := []string{"Title", msg}
	for _, value := range params {
		valStr := fmt.Sprintf("%v", value)
		valStr = strings.ReplaceAll(valStr, "\"", "'")
		valStr = strings.ReplaceAll(valStr, "\n", "\\n")
		slice = append(slice, valStr)
	}
	//slice = append(slice, "Now", util.UnixToSimpleStr(time.Now().Unix()))
	slice = append(slice, "SinceStart", time.Since(START_TIME).String())

	var b strings.Builder
	for i := 0; i < len(slice); i += 2 {
		b.WriteString(fmt.Sprintf("%s: %s\\n", slice[i], slice[i+1]))
	}
	return b.String(), nil
}
