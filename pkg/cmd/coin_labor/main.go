package main

import (
	"flag"
	"fmt"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/util"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var configFile = flag.String("config", "", "path to config file")
var appConfigFile = flag.String("app-config", "", "path to app config file")
var pidFile = flag.String("pidfile", "", "path to pid file")

func main() {
	flag.Parse()

	if *configFile == "" {
		*configFile = "conf/dev.ini"
	}

	fmt.Println("starting... time: " + util.UnixToStr(time.Now().Unix()))
	server := NewLaborServer()

	go listenToSystemSignals(server)

	err := server.Run()

	time.Sleep(1 * time.Second)
	code := server.Exit(err)
	log.Close()

	fmt.Println("stopped time: " + util.UnixToStr(time.Now().Unix()))
	os.Exit(code)
}

func listenToSystemSignals(server *LaborServerImpl) {
	signalChan := make(chan os.Signal, 1)
	sighupChan := make(chan os.Signal, 1)

	signal.Notify(sighupChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case sig := <-sighupChan:
			fmt.Printf("System signal: %s, type: SIGHUP", sig)
			log.Reload()
		case sig := <-signalChan:
			go func() {
				// force kill
				sec := 10
				ticker := time.NewTicker(time.Duration(sec) * time.Second)
				<-ticker.C
				log.New("OS").Error("force exited.", "wait second", sec)
				os.Exit(1)
			}()
			server.Shutdown(fmt.Sprintf("System signal: %s", sig))
			return
		}
	}
}
