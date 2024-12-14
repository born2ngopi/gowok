package config

import (
	"io"
	"log/slog"
	"os"

	"github.com/ngamux/middleware/cors"
	"github.com/ngamux/middleware/log"
	"github.com/ngamux/middleware/pprof"
)

type App struct {
	Key  string
	Web  Web
	Grpc Grpc
}

type Web struct {
	Enabled bool
	Host    string

	Log *struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"log"`

	Cors *struct {
		AllowOrigins     string `yaml:"allow_origins"`
		AllowCredentials bool   `yaml:"allow_credentials"`
		AllowMethods     string `yaml:"allow_methods"`
		AllowHeaders     string `yaml:"allow_headers"`
		MaxAge           int    `yaml:"max_age"`
		ExposeHeaders    string `yaml:"expose_headers"`
	} `yaml:"cors"`

	Pprof *struct {
		Enabled bool   `yaml:"enabled"`
		Prefix  string `yaml:"prefix"`
	} `yaml:"pprof"`

	Views  WebViews  `yaml:"views"`
	Static WebStatic `yaml:"static"`
}

type WebViews struct {
	Enabled bool   `yaml:"enabled"`
	Dir     string `yaml:"dir"`
	Layout  string `yaml:"layout"`
}

type WebStatic struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
	Dir     string `yaml:"dir"`
}

func (r Web) GetLog() log.Config {
	c := log.Config{
		Handler: slog.NewTextHandler(io.Discard, nil),
	}
	if r.Log == nil {
		return c
	}

	if r.Log.Enabled {
		c.Handler = slog.NewJSONHandler(os.Stdout, nil)
	}
	return c
}

func (r Web) GetCors() cors.Config {
	c := cors.Config{}
	if r.Cors == nil {
		return c
	}
	if r.Cors.AllowOrigins != "" {
		c.AllowOrigins = r.Cors.AllowOrigins
	}
	if r.Cors.AllowMethods != "" {
		c.AllowMethods = r.Cors.AllowMethods
	}
	if r.Cors.AllowHeaders != "" {
		c.AllowHeaders = r.Cors.AllowHeaders
	}
	// if r.Cors.ExposeHeaders != "" {
	// 	c.ExposeHeaders = r.Cors.ExposeHeaders
	// }
	// if r.Cors.AllowCredentials != false {
	// 	c.AllowCredentials = r.Cors.AllowCredentials
	// }
	// if r.Cors.MaxAge != 0 {
	// 	c.MaxAge = r.Cors.MaxAge
	// }
	return c
}

func (r Web) GetPprof() pprof.Config {
	c := pprof.Config{}
	if r.Pprof == nil {
		return c
	}

	if r.Pprof.Prefix != "" {
		c.Prefix = r.Pprof.Prefix
	}
	return c
}

func (r Web) GetViews() WebViews {
	v := WebViews{
		Enabled: r.Views.Enabled,
		Layout:  r.Views.Layout,
	}
	if !v.Enabled {
		return v
	}
	if r.Views.Dir == "" {
		v.Dir = "./views"
	}
	return v
}

func (r Web) GetStatic() WebStatic {
	v := WebStatic{
		Enabled: r.Static.Enabled,
		Dir:     r.Static.Dir,
		Prefix:  "/public",
	}
	if !v.Enabled {
		return v
	}
	if r.Static.Dir == "" {
		v.Dir = "./public"
	}
	if r.Static.Prefix != "" {
		v.Prefix = r.Static.Prefix
	}
	return v
}

type Grpc struct {
	Enabled bool
	Host    string
}
