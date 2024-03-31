package setting

import (
	"bytes"
	"fmt"
	"gopkg.in/ini.v1"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"jasonzhu.com/coin_labor/core/components/log"
)

/**

@author Jason
@version 2020-06-25 19:28
*/

type Scheme string

const (
	DEFAULT_CONFIG_FILE = "conf/defaults.ini"

	REDACTED_VALUE = "*********"
)

const (
	APP_NAME = "Labor"

	DEV  = "development"
	PROD = "production"
)

var (
	// logger
	lg log.Logger

	// Raw Global setting objects.
	Raw *ini.File

	// for logging purposes
	configFiles                  []string
	appliedCommandLineProperties []string
	appliedEnvOverrides          []string
)

var (
	// App settings.
	Env             = DEV
	ApplicationName = APP_NAME

	// Paths
	HomePath       string
	DataPath       string
	LogsPath       string
	CustomInitPath = "conf/custom.ini"

	// Http server options

	// Alerting
	AlertingEnabled bool
	TgToken         string
)

type Cfg struct {
}

type CommandLineArgs struct {
	Config    string
	AppConfig string
	HomePath  string
	Args      []string
}

func init() {
	lg = log.New("settings")
}

func shouldRedactKey(s string) bool {
	upperCased := strings.ToUpper(s)
	return strings.Contains(upperCased, "PASSWORD") ||
		strings.Contains(upperCased, "SECRET") ||
		strings.Contains(upperCased, "PROVIDER_CONFIG")
}

func shouldRedactURLKey(s string) bool {
	upperCased := strings.ToUpper(s)
	return strings.Contains(upperCased, "DATABASE_URL")
}

func applyEnvVariableOverrides(file *ini.File) error {
	appliedEnvOverrides = make([]string, 0)
	for _, section := range file.Sections() {
		for _, key := range section.Keys() {
			sectionName := strings.ToUpper(strings.Replace(section.Name(), ".", "_", -1))
			keyName := strings.ToUpper(strings.Replace(key.Name(), ".", "_", -1))
			envKey := fmt.Sprintf("GF_%s_%s", sectionName, keyName)
			envValue := os.Getenv(envKey)

			if len(envValue) > 0 {
				key.SetValue(envValue)
				if shouldRedactKey(envKey) {
					envValue = REDACTED_VALUE
				}

				if shouldRedactURLKey(envKey) {
					u, err := url.Parse(envValue)
					if err != nil {
						return fmt.Errorf("could not parse environment variable. key: %s, value: %s. error: %v", envKey, envValue, err)
					}
					ui := u.User
					if ui != nil {
						_, exists := ui.Password()
						if exists {
							u.User = url.UserPassword(ui.Username(), "-redacted-")
							envValue = u.String()
						}
					}
				}
				appliedEnvOverrides = append(appliedEnvOverrides, fmt.Sprintf("%s=%s", envKey, envValue))
			}
		}
	}

	return nil
}

func applyCommandLineDefaultProperties(props map[string]string, file *ini.File) {
	appliedCommandLineProperties = make([]string, 0)
	for _, section := range file.Sections() {
		for _, key := range section.Keys() {
			keyString := fmt.Sprintf("default.%s.%s", section.Name(), key.Name())
			value, exists := props[keyString]
			if exists {
				key.SetValue(value)
				if shouldRedactKey(keyString) {
					value = REDACTED_VALUE
				}
				appliedCommandLineProperties = append(appliedCommandLineProperties, fmt.Sprintf("%s=%s", keyString, value))
			}
		}
	}
}

func applyCommandLineProperties(props map[string]string, file *ini.File) {
	for _, section := range file.Sections() {
		sectionName := section.Name() + "."
		if section.Name() == ini.DEFAULT_SECTION {
			sectionName = ""
		}
		for _, key := range section.Keys() {
			keyString := sectionName + key.Name()
			value, exists := props[keyString]
			if exists {
				appliedCommandLineProperties = append(appliedCommandLineProperties, fmt.Sprintf("%s=%s", keyString, value))
				key.SetValue(value)
			}
		}
	}
}

func getCommandLineProperties(args []string) map[string]string {
	props := make(map[string]string)

	for _, arg := range args {
		if !strings.HasPrefix(arg, "cfg:") {
			continue
		}

		trimmed := strings.TrimPrefix(arg, "cfg:")
		parts := strings.Split(trimmed, "=")
		if len(parts) != 2 {
			log.Fatal("Invalid command line argument. argument: %v", arg)
			return nil
		}

		props[parts[0]] = parts[1]
	}
	return props
}

func makeAbsolute(path string, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

func evalEnvVarExpression(value string) string {
	regex := regexp.MustCompile(`\${(\w+)}`)
	return regex.ReplaceAllStringFunc(value, func(envVar string) string {
		envVar = strings.TrimPrefix(envVar, "${")
		envVar = strings.TrimSuffix(envVar, "}")
		envValue := os.Getenv(envVar)

		// if env variable is hostname and it is empty use os.Hostname as default
		if envVar == "HOSTNAME" && envValue == "" {
			envValue, _ = os.Hostname()
		}

		return envValue
	})
}

func evalConfigValues(file *ini.File) {
	for _, section := range file.Sections() {
		for _, key := range section.Keys() {
			key.SetValue(evalEnvVarExpression(key.Value()))
		}
	}
}

func loadSpecifiedConfigFile(configFile string, masterFile *ini.File) error {
	if configFile == "" {
		configFile = filepath.Join(HomePath, CustomInitPath)
		// return without error if custom file does not exist
		if !pathExists(configFile) {
			return nil
		}
	}

	userConfig, err := ini.Load(configFile)
	if err != nil {
		return fmt.Errorf("Failed to parse %v, %v", configFile, err)
	}

	userConfig.BlockMode = false

	for _, section := range userConfig.Sections() {
		for _, key := range section.Keys() {
			if key.Value() == "" {
				continue
			}

			defaultSec, err := masterFile.GetSection(section.Name())
			if err != nil {
				defaultSec, _ = masterFile.NewSection(section.Name())
			}
			defaultKey, err := defaultSec.GetKey(key.Name())
			if err != nil {
				defaultKey, _ = defaultSec.NewKey(key.Name(), key.Value())
			}
			defaultKey.SetValue(key.Value())
		}
	}

	configFiles = append(configFiles, configFile)
	return nil
}

func (cfg *Cfg) loadConfiguration(args *CommandLineArgs) (*ini.File, error) {
	var err error

	//load config defaults
	defaultConfigFile := path.Join(HomePath, DEFAULT_CONFIG_FILE)
	configFiles = append(configFiles, defaultConfigFile)

	// check if config file exists
	if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
		fmt.Println("Pipe-server Init Failed: Could not find config defaults, make sure homepath command line parameter is set or working directory is homepath")
		os.Exit(1)
	}

	// load defaults
	parsedFile, err := ini.Load(defaultConfigFile)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to parse defaults.ini, %v", err))
		os.Exit(1)
		return nil, err
	}

	parsedFile.BlockMode = false

	// command line props
	commandLineProps := getCommandLineProperties(args.Args)
	// load default overrides
	applyCommandLineDefaultProperties(commandLineProps, parsedFile)

	// load specified config file
	err = loadSpecifiedConfigFile(args.Config, parsedFile)
	if err != nil {
		cfg.initLogging(parsedFile)
		log.Fatal(err.Error())
	}

	// apply environment overrides
	err = applyEnvVariableOverrides(parsedFile)
	if err != nil {
		return nil, err
	}

	// apply command line overrides
	applyCommandLineProperties(commandLineProps, parsedFile)

	// evaluate config values containing environment variables
	evalConfigValues(parsedFile)

	// update data path and logging config
	DataPath = makeAbsolute(parsedFile.Section("paths").Key("data").String(), HomePath)
	cfg.initLogging(parsedFile)

	return parsedFile, err
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func setHomePath(args *CommandLineArgs) {
	if args.HomePath != "" {
		HomePath = args.HomePath
		return
	}

	HomePath, _ = filepath.Abs(".")
	// check if homePath is correct
	if pathExists(filepath.Join(HomePath, DEFAULT_CONFIG_FILE)) {
		return
	}

	// try down one path
	if pathExists(filepath.Join(HomePath, "../", DEFAULT_CONFIG_FILE)) {
		HomePath = filepath.Join(HomePath, "../")
	}
}

func NewCfg() *Cfg {
	return &Cfg{}
}

func (cfg *Cfg) Load(args *CommandLineArgs) error {
	setHomePath(args)

	iniFile, err := cfg.loadConfiguration(args)
	if err != nil {
		return err
	}

	// Temporary keep global, to make refactor in steps
	Raw = iniFile

	Env = iniFile.Section("").Key("app_mode").MustString("development")

	alerting := iniFile.Section("alerting")
	AlertingEnabled = alerting.Key("enabled").MustBool(true)
	TgToken = alerting.Key("telegram_token").String()

	return nil
}

func (cfg *Cfg) initLogging(file *ini.File) {
	// split on comma
	logModes := strings.Split(file.Section("log").Key("mode").MustString("console"), ",")
	// also try space
	if len(logModes) == 1 {
		logModes = strings.Split(file.Section("log").Key("mode").MustString("console"), " ")
	}
	LogsPath = makeAbsolute(file.Section("paths").Key("logs").String(), HomePath)
	log.ReadLoggingConfig(logModes, LogsPath, file)
}

func (cfg *Cfg) LogConfigSources() {
	var text bytes.Buffer

	for _, file := range configFiles {
		lg.Info("Configuration loaded", "file", file)
	}

	if len(appliedCommandLineProperties) > 0 {
		for _, prop := range appliedCommandLineProperties {
			lg.Info("Config overridden from command line", "arg", prop)
		}
	}

	if len(appliedEnvOverrides) > 0 {
		text.WriteString("\tEnvironment variables used:\n")
		for _, prop := range appliedEnvOverrides {
			lg.Info("Config overridden from Environment variable", "var", prop)
		}
	}

	lg.Info("Path Home", "path", HomePath)
	lg.Info("Path Data", "path", DataPath)
	lg.Info("Path Logs", "path", LogsPath)
	lg.Info("Application mode (ENV) = " + Env)
}
