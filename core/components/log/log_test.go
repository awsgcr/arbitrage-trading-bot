package log

import (
	"fmt"
	"testing"
	"time"
)

func TestLogfmtFormat(t *testing.T) {
	fmt.Println(time.Now().Format(timeFormat))
}
