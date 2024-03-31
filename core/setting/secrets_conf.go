package setting

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/util/homedir"
	"path/filepath"
)

const (
	storePath        = "~/.coin_labor"
	secretConfigPath = "conf/secrets.conf.yml"
)

var SecretsConf *SecretCfg

func init() {
	config := newSecretConfig()
	err := config.LoadAppConfiguration()
	if err != nil {
		log.Fatal("failed to load secrets configuration: %v", err)
	}
	SecretsConf = config
}

func newSecretConfig() *SecretCfg {
	return &SecretCfg{}
}

type SecretCfg struct {
	configFile string

	Binance Secret `yaml:"binance"`
	MEXC    Secret `yaml:"mexc"`
}

type Secret struct {
	Key    string `yaml:"key"`
	Secret string `yaml:"secret"`
}

func (cfg *SecretCfg) LoadAppConfiguration() error {
	path := secretConfigPath
	expandedDir, err := homedir.Expand(storePath)
	if err != nil {
		return err
	}
	cfg.configFile = filepath.Join(expandedDir, path)
	if !pathExists(cfg.configFile) {
		return errors.New("could not found secret config file, path: " + cfg.configFile)
	}

	if err := cfg.loadFromFile(); err != nil {
		return errors.New(fmt.Sprintf("failed to parse secret config file: %v", err))
	}
	return nil
}

func (cfg *SecretCfg) loadFromFile() error {
	b, err := ioutil.ReadFile(cfg.configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return err
	}

	return nil
}
