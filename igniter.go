package gowok

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gowok/gowok/grpc"
	"github.com/gowok/gowok/health"
	"github.com/gowok/gowok/must"
	"github.com/gowok/gowok/router"
	"github.com/gowok/gowok/runner"
	"github.com/gowok/gowok/singleton"
	"github.com/gowok/gowok/sql"
)

type ConfigureFunc func(*Project)

type Project struct {
	Config     *Config
	ConfigMap  map[string]any
	Runner     *runner.Runner
	Hooks      *Hooks
	configures []ConfigureFunc
}

var project *Project

func ignite() (*Project, error) {
	var pathConfig string
	if flag.Lookup("config") == nil {
		flag.StringVar(&pathConfig, "config", "config.yaml", "configuration file location (yaml)")
	} else {
		pathConfig = flag.Lookup("config").Value.String()
	}
	flag.Parse()

	conf, confRaw, err := NewConfig(pathConfig)
	if err != nil {
		return nil, err
	}

	hooks := &Hooks{}
	running := runner.New(
		runner.WithRLimitEnable(true),
		runner.WithGracefulStopFunc(func() {
			println()
			if conf.App.Grpc.Enabled {
				slog.Info("stopping GRPC")
				grpc.Server().GracefulStop()
			}
			if conf.App.Web.Enabled {
				slog.Info("stopping web")
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				_ = router.Server().Shutdown(ctx)
			}
			hooks.onStopped.IfPresent(func(f Hook) {
				f()
			})
		}),
	)

	project = &Project{
		Config:     conf,
		ConfigMap:  confRaw,
		Runner:     running,
		Hooks:      hooks,
		configures: make([]ConfigureFunc, 0),
	}
	project.Configures(func(p *Project) {
		sql.Configure(p.Config.SQLs)
		router.Configure(&p.Config.App.Web)
		grpc.Configure(&p.Config.App.Grpc)
		health.Configure()
	})

	return project, nil
}

var projectSingleton = singleton.New(func() *Project {
	return must.Must(ignite())
})

func Get() *Project {
	pp := projectSingleton()
	return *pp
}

func run(project *Project) {
	project.Hooks.onStarting.IfPresent(func(f Hook) {
		f()
	})

	go func() {
		if !project.Config.App.Web.Enabled {
			return
		}

		slog.Info("starting web")
		err := router.Server().ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			panic("web can't start, because: " + err.Error())
		}
	}()

	go func() {
		if !project.Config.App.Grpc.Enabled {
			return
		}

		slog.Info("starting GRPC")
		listen, err := net.Listen("tcp", project.Config.App.Grpc.Host)
		if err != nil {
			panic("GRPC can't start, because: " + err.Error())
		}

		err = grpc.Server().Serve(listen)
		if err != nil {
			panic("GRPC can't start, because: " + err.Error())
		}
	}()

	project.Hooks.onStarted.IfPresent(func(f Hook) {
		f()
	})
}

func (p *Project) Run(forever ...bool) {
	p.Runner.AddRunFunc(func() {
		run(p)
	})
	if p.Config.App.Web.Enabled || p.Config.App.Grpc.Enabled {
		forever = append([]bool{true}, forever...)
	}
	p.Runner.Run(forever...)
}

func (p *Project) Configures(configures ...ConfigureFunc) *Project {
	p.configures = append(p.configures, configures...)
	for _, configure := range configures {
		configure(project)
	}
	return p
}

func (p *Project) Reload() {
	must.Must(ignite())
}
