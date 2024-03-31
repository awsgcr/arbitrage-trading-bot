package log

import (
	"fmt"
	"gopkg.in/ini.v1"
	"jasonzhu.com/coin_labor/core/util"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-stack/stack"
	"github.com/inconshreveable/log15"
	"github.com/mattn/go-isatty"
)

/**

@author Jason
@version 2020-06-25 19:45
*/

var Root log15.Logger
var loggersToClose []DisposableHandler
var loggersToReload []ReloadableHandler
var filters map[string]log15.Lvl

func init() {
	loggersToClose = make([]DisposableHandler, 0)
	loggersToReload = make([]ReloadableHandler, 0)
	filters = map[string]log15.Lvl{}
	Root = log15.Root()
	Root.SetHandler(log15.DiscardHandler())
}

func New(logger string, ctx ...interface{}) Logger {
	params := append([]interface{}{"logger", logger}, ctx...)
	return Root.New(params...)
}

func Trace(format string, v ...interface{}) {
	var message string
	if len(v) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = format
	}

	Root.Debug(message)
}

func Debug(format string, v ...interface{}) {
	var message string
	if len(v) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = format
	}

	Root.Debug(message)
}

func Info(format string, v ...interface{}) {
	var message string
	if len(v) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = format
	}

	Root.Info(message)
}

func Warn(format string, v ...interface{}) {
	var message string
	if len(v) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = format
	}

	Root.Warn(message)
}

func Error(format string, v ...interface{}) {
	Root.Error(fmt.Sprintf(format, v...))
}

func Critical(format string, v ...interface{}) {
	Root.Crit(fmt.Sprintf(format, v...))
}

func Fatal(format string, v ...interface{}) {
	fmt.Printf(format, v...)
	Root.Crit(fmt.Sprintf(format, v...))
	Close()
	os.Exit(1)
}

func Close() {
	for _, logger := range loggersToClose {
		logger.Close()
	}
	loggersToClose = make([]DisposableHandler, 0)
}

func Reload() {
	for _, logger := range loggersToReload {
		logger.Reload()
	}
}

func GetLogLevelFor(name string) Lvl {
	if level, ok := filters[name]; ok {
		switch level {
		case log15.LvlWarn:
			return LvlWarn
		case log15.LvlInfo:
			return LvlInfo
		case log15.LvlError:
			return LvlError
		case log15.LvlCrit:
			return LvlCrit
		default:
			return LvlDebug
		}
	}

	return LvlInfo
}

var logLevels = map[string]log15.Lvl{
	"trace":    log15.LvlDebug,
	"debug":    log15.LvlDebug,
	"info":     log15.LvlInfo,
	"warn":     log15.LvlWarn,
	"error":    log15.LvlError,
	"critical": log15.LvlCrit,
}

func getLogLevelFromConfig(key string, defaultName string, cfg *ini.File) (string, log15.Lvl) {
	levelName := cfg.Section(key).Key("level").MustString(defaultName)
	levelName = strings.ToLower(levelName)
	level := getLogLevelFromString(levelName)
	return levelName, level
}

func getLogLevelFromString(levelName string) log15.Lvl {
	level, ok := logLevels[levelName]

	if !ok {
		Root.Error("Unknown log level", "level", levelName)
		return log15.LvlError
	}

	return level
}

func getFilters(filterStrArray []string) map[string]log15.Lvl {
	filterMap := make(map[string]log15.Lvl)

	for _, filterStr := range filterStrArray {
		parts := strings.Split(filterStr, ":")
		if len(parts) > 1 {
			filterMap[parts[0]] = getLogLevelFromString(parts[1])
		}
	}

	return filterMap
}

func getLogFormat(format string) log15.Format {
	switch format {
	case "console":
		if isatty.IsTerminal(os.Stdout.Fd()) {
			return log15.TerminalFormat()
		}
		return LogfmtFormat()
	case "text":
		return LogfmtFormat()
	case "json":
		return JsonFormat()
	default:
		return LogfmtFormat()
	}
}

func ReadLoggingConfig(modes []string, logsPath string, cfg *ini.File) {
	Close()

	defaultLevelName, _ := getLogLevelFromConfig("log", "info", cfg)
	defaultFilters := getFilters(util.SplitString(cfg.Section("log").Key("filters").String()))

	handlers := make([]log15.Handler, 0)

	for _, mode := range modes {
		mode = strings.TrimSpace(mode)
		sec, err := cfg.GetSection("log." + mode)
		if err != nil {
			Root.Error("Unknown log mode", "mode", mode)
		}

		// Log level.
		_, level := getLogLevelFromConfig("log."+mode, defaultLevelName, cfg)
		modeFilters := getFilters(util.SplitString(sec.Key("filters").String()))
		format := getLogFormat(sec.Key("format").MustString(""))

		var handler log15.Handler

		// Generate log configuration.
		switch mode {
		case "console":
			handler = log15.StreamHandler(os.Stdout, format)
		case "file":
			fileName := sec.Key("file_name").MustString(filepath.Join(logsPath, "pipe.log"))
			_ = os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
			fileHandler := NewFileWriter()
			fileHandler.Filename = fileName
			fileHandler.Format = format
			fileHandler.Rotate = sec.Key("log_rotate").MustBool(true)
			fileHandler.Maxlines = sec.Key("max_lines").MustInt(1000000)
			fileHandler.Maxsize = 1 << uint(sec.Key("max_size_shift").MustInt(28))
			fileHandler.Daily = sec.Key("daily_rotate").MustBool(true)
			fileHandler.Maxdays = sec.Key("max_days").MustInt64(7)
			fileHandler.Init()

			loggersToClose = append(loggersToClose, fileHandler)
			loggersToReload = append(loggersToReload, fileHandler)
			handler = fileHandler
		}

		for key, value := range defaultFilters {
			if _, exist := modeFilters[key]; !exist {
				modeFilters[key] = value
			}
		}

		for key, value := range modeFilters {
			if _, exist := filters[key]; !exist {
				filters[key] = value
			}
		}

		handler = LogFilterHandler(level, modeFilters, handler)
		handlers = append(handlers, handler)
	}

	Root.SetHandler(log15.MultiHandler(handlers...))
}

func LogFilterHandler(maxLevel log15.Lvl, filters map[string]log15.Lvl, h log15.Handler) log15.Handler {
	return log15.FilterHandler(func(r *log15.Record) (pass bool) {

		if len(filters) > 0 {
			for i := 0; i < len(r.Ctx); i += 2 {
				key, ok := r.Ctx[i].(string)
				if ok && key == "logger" {
					loggerName, strOk := r.Ctx[i+1].(string)
					if strOk {
						if filterLevel, ok := filters[loggerName]; ok {
							return r.Lvl <= filterLevel
						}
					}
				}
			}
		}

		return r.Lvl <= maxLevel
	}, h)
}

func Stack(skip int) string {
	call := stack.Caller(skip)
	s := stack.Trace().TrimBelow(call).TrimRuntime()
	return s.String()
}
