package gowok

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/gowok/gowok/driver"
	"github.com/gowok/gowok/must"
	"github.com/gowok/gowok/optional"
	"github.com/gowok/gowok/runner"
	"github.com/ngamux/ngamux"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

type getterByName[T any] func(name ...string) optional.Optional[T]
type ConfigureFunc func(*Project)

type Project struct {
	Config     *Config
	Runner     *runner.Runner
	Hooks      *Hooks
	SQL        getterByName[*sql.DB]
	MongoDB    getterByName[*mongo.Client]
	Cache      getterByName[*cache.Cache[any]]
	Validator  *Validator
	Web        *ngamux.HttpServeMux
	webServer  *HttpMux
	GRPC       *grpc.Server
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

	conf, err := NewConfig(pathConfig)
	if err != nil {
		return nil, err
	}

	dbSQL, err := driver.NewSQL(conf.SQLs)
	if err != nil {
		return nil, err
	}

	dbMongo, err := driver.NewMongoDB(conf.MongoDBs)
	if err != nil {
		return nil, err
	}

	dbCache, err := driver.NewCache(conf.Caches)
	if err != nil {
		return nil, err
	}

	validator := NewValidator()

	web := NewHTTP(&conf.App.Web)
	GRPC := grpc.NewServer()

	hooks := &Hooks{}
	running := runner.New(
		runner.WithRLimitEnable(true),
		runner.WithGracefulStopFunc(func() {
			println()
			if conf.App.Grpc.Enabled {
				println("project: stopping GRPC")
				GRPC.GracefulStop()
			}
			if conf.App.Web.Enabled {
				println("project: stopping web")
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				web.Server.Shutdown(ctx)
			}
			if hooks.onStopped != nil {
				hooks.onStopped()
			}
		}),
		runner.WithRunFunc(run),
	)
	project = &Project{
		Config:    conf,
		Runner:    running,
		Hooks:     hooks,
		SQL:       dbSQL.Get,
		MongoDB:   dbMongo.Get,
		Cache:     dbCache.Get,
		Validator: validator,
		Web:       web.Mux,
		webServer: web,
		GRPC:      GRPC,

		configures: make([]ConfigureFunc, 0),
	}
	return project, nil
}

func Get() *Project {
	if project != nil {
		return project
	}

	must.Must(ignite())
	return project
}

func run() {
	project := Get()

	for _, configure := range project.configures {
		configure(project)
	}

	if project.Hooks.onStarting != nil {
		project.Hooks.onStarting()
	}

	go func() {
		if !project.Config.App.Web.Enabled {
			return
		}

		println("project: starting web")
		err := project.webServer.Server.ListenAndServe()
		if err != nil {
			log.Fatalf("web can't start, because: %v", err)
		}
	}()

	go func() {
		if !project.Config.App.Grpc.Enabled {
			return
		}

		println("project: starting GRPC")
		listen, err := net.Listen("tcp", project.Config.App.Grpc.Host)
		if err != nil {
			log.Fatalf("GRPC can't start, because: %v", err)
		}

		err = project.GRPC.Serve(listen)
		if err != nil {
			log.Fatalf("GRPC can't start, because: %v", err)
		}
	}()

	if project.Hooks.onStarted != nil {
		project.Hooks.onStarted()
	}
}

func (p *Project) Run() {
	Get().Runner.Run()
}

func (p *Project) Configures(configures ...ConfigureFunc) {
	p.configures = make([]ConfigureFunc, len(configures))
	copy(p.configures, configures)
}

func (p *Project) Reload() {
	must.Must(ignite())
}
