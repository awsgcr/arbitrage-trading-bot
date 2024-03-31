package setting

import (
	"fmt"
	"testing"
)

func TestNewSecretConfig(t *testing.T) {
	//cfg := setting.NewCfg()
	//cfg.Load(&setting.CommandLineArgs{
	//	HomePath: "/Users/jasonzjs/Workspaces_labor/coin_labor",
	//})

	config := newSecretConfig()
	err := config.LoadAppConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(config)
}
