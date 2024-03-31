package main

import (
	"context"
	"fmt"
	"github.com/facebookgo/inject"
	"golang.org/x/sync/errgroup"

	"jasonzhu.com/coin_labor/core/components/bus"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/components/registry"
	"jasonzhu.com/coin_labor/core/setting"

	_ "jasonzhu.com/coin_labor/pkg/plugins"
	_ "jasonzhu.com/coin_labor/pkg/services"
	//_ "jasonzhu.com/coin_labor/pkg/services/trader"
)

func NewLaborServer() *LaborServerImpl {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	return &LaborServerImpl{
		context:       childCtx,
		shutdownFn:    shutdownFn,
		childRoutines: childRoutines,
		log:           log.New("server"),
		cfg:           setting.NewCfg(),
	}
}

type LaborServerImpl struct {
	context            context.Context
	shutdownFn         context.CancelFunc
	childRoutines      *errgroup.Group
	log                log.Logger
	cfg                *setting.Cfg
	shutdownReason     string
	shutdownInProgress bool
}

func (g *LaborServerImpl) Run() (err error) {
	g.loadConfiguration()
	g.writePIDFile()

	serviceGraph := inject.Graph{}
	err = serviceGraph.Provide(&inject.Object{Value: bus.GetBus()})
	if err != nil {
		return fmt.Errorf("failed to provide object to the graph: %v", err)
	}

	// self registered services
	services := registry.GetServices()

	// Add all services to dependency graph
	for _, service := range services {
		err = serviceGraph.Provide(&inject.Object{Value: service.Instance})
		if err != nil {
			return fmt.Errorf("failed to provide object to the graph: %v", err)
		}
	}

	err = serviceGraph.Provide(&inject.Object{Value: g})
	if err != nil {
		return fmt.Errorf("failed to provide object to the graph: %v", err)
	}

	// Inject dependencies to services
	if err := serviceGraph.Populate(); err != nil {
		return fmt.Errorf("failed to populate service dependency: %v", err)
	}

	// Init & start services
	for _, service := range services {
		if registry.IsDisabled(service.Instance) {
			continue
		}

		g.log.Info("Initializing " + service.Name)

		if err := service.Instance.Init(); err != nil {
			return fmt.Errorf("service init failed: %v", err)
		}
	}
	g.log.Info("All services Initialized.")

	// Start background services
	for _, srv := range services {
		// variable needed for accessing loop variable in function callback
		descriptor := srv
		service, ok := srv.Instance.(registry.BackgroundService)
		if !ok {
			continue
		}

		if registry.IsDisabled(descriptor.Instance) {
			continue
		}

		g.childRoutines.Go(func() error {
			// Skip starting new service when shutting down
			// Can happen when service stop/return during startup
			if g.shutdownInProgress {
				return nil
			}

			g.log.Info("Background Running: " + descriptor.Name)

			err := service.Run(g.context)

			// If error is not canceled then the service crashed
			if err != context.Canceled && err != nil {
				g.log.Error("Stopped "+descriptor.Name, "reason", err)
			} else {
				g.log.Info("Finished "+descriptor.Name, "reason", err)
			}

			return err
		})
	}
	g.log.Info("Starting background services", "len", len(services))

	_ = sendSystemdNotification("READY=1")

	return g.childRoutines.Wait()
}
