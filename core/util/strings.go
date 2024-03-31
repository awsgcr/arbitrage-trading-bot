package util

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// StringsFallback2 returns the first of two not empty strings.
func StringsFallback2(val1 string, val2 string) string {
	return stringsFallback(val1, val2)
}

// StringsFallback3 returns the first of three not empty strings.
func StringsFallback3(val1 string, val2 string, val3 string) string {
	return stringsFallback(val1, val2, val3)
}

func stringsFallback(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// SplitString splits a string by commas or empty spaces.
func SplitString(str string) []string {
	if len(str) == 0 {
		return []string{}
	}

	return regexp.MustCompile("[, ]+").Split(str, -1)
}

// GetAgeString returns a string representing certain time from years to minutes.
func GetAgeString(t time.Time) string {
	if t.IsZero() {
		return "?"
	}

	sinceNow := time.Since(t)
	minutes := sinceNow.Minutes()
	years := int(math.Floor(minutes / 525600))
	months := int(math.Floor(minutes / 43800))
	days := int(math.Floor(minutes / 1440))
	hours := int(math.Floor(minutes / 60))

	if years > 0 {
		return fmt.Sprintf("%dy", years)
	}
	if months > 0 {
		return fmt.Sprintf("%dM", months)
	}
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	if int(minutes) > 0 {
		return fmt.Sprintf("%dm", int(minutes))
	}

	return "< 1m"
}

func ToString(v interface{}) (string, error) {
	var vStr string
	switch v.(type) {
	case string:
		vStr = v.(string)
	case float32, float64:
		vStr = strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', 10, 64)
	case int, int8, int16, int32, int64:
		vStr = strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		vStr = strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	case nil:
		return "", nil
	default:
		err := fmt.Errorf("type not supported")
		return "", err
	}
	return strings.TrimSpace(vStr), nil
}
