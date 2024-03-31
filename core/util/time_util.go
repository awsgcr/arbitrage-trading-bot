package util

import (
	"strings"
	"time"
)

/**

@author Jason
@version 2019-05-07 17:57
*/
const (
	LayoutDate         = "2006-01-02"
	LayoutDateTime     = "2006-01-02 15:04:05"
	LayoutDateMillTime = "2006-01-02 15:04:05.000"
)

type TimeDTO struct {
	DateTime string
	Unix     int64
}

func (t *TimeDTO) AddUnix(second int64) *TimeDTO {
	n := t.Unix + second
	ns := UnixToStr(n)
	return &TimeDTO{
		DateTime: ns,
		Unix:     n,
	}
}

func FromDateTimeStr(str string) (*TimeDTO, error) {
	unix, e := GetUnixOfDateTime(str)
	if e != nil {
		return nil, e
	}
	return &TimeDTO{
		DateTime: str,
		Unix:     unix,
	}, nil
}

func GetUnixOfDateTime(str string) (int64, error) {
	location, e := ParseDateTimeStr(str)
	if e != nil {
		return 0, e
	}
	return location.Unix(), nil
}
func GetMsOfDateTime(str string) (int64, error) {
	location, e := ParseDateTimeStr(str)
	if e != nil {
		return 0, e
	}
	return location.UnixNano() / 1000000, nil
}
func ParseDateTimeStr(str string) (*time.Time, error) {
	location, e := time.ParseInLocation(LayoutDateTime, str, time.Local)
	if e != nil {
		return nil, e
	}
	return &location, nil
}

func UnixToStr(unix int64) string {
	format := time.Unix(unix, 0).Format(LayoutDateTime)
	return format
}

func UnixMillToStr(msec int64) string {
	return time.UnixMilli(msec).Format(LayoutDateMillTime)
}

// NowTimeString now, today
func NowTimeString() string {
	t, _ := TimeParse("now")
	return t.DateTime
}

func TimeParse(s string) (*TimeDTO, error) {
	now := time.Now()
	switch strings.ToLower(s) {
	case "now":
		return &TimeDTO{
			DateTime: now.Format(LayoutDateTime),
			Unix:     now.Unix(),
		}, nil
	case "today":
		timeStr := now.Format(LayoutDate)
		t, _ := time.ParseInLocation(LayoutDateTime, timeStr+" 00:00:00", time.Local)
		return &TimeDTO{
			DateTime: t.Format(LayoutDateTime),
			Unix:     t.Unix(),
		}, nil
	case "yesterday":
		timeStr := now.AddDate(0, 0, -1).Format(LayoutDate)
		t, _ := time.ParseInLocation(LayoutDateTime, timeStr+" 00:00:00", time.Local)
		return &TimeDTO{
			DateTime: t.Format(LayoutDateTime),
			Unix:     t.Unix(),
		}, nil
	default:
		return FromDateTimeStr(s)
	}
}
