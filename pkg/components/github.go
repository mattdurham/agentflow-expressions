package components

import (
	"agentflow/pkg/config"
	"context"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	gh_config "github.com/infinityworks/github-exporter/config"
	"github.com/infinityworks/github-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
)

type Github struct {
	ex        *exporter.Exporter
	configMut sync.Mutex
	cfg       *config.Github
	reloadCfg chan struct{}
	log       log.Logger
	registry  *prometheus.Registry
}

func NewGithub(cfg config.Github, log log.Logger) (*Github, error) {
	ghExporter,r, err := newGithubExporter(cfg, log)
	if err != nil {
		return nil, err
	}
	return &Github{
		ex:        ghExporter,
		reloadCfg: make(chan struct{}, 1),
		log:       log,
		registry: r,
	}, nil
}

func newGithubExporter(cfg config.Github, log log.Logger) (*exporter.Exporter,*prometheus.Registry, error) {
	if cfg.ApiURL == "" {
		cfg.ApiURL = "https://api.github.com"
	}
	conf := gh_config.Config{}
	err := conf.SetAPIURL(cfg.ApiURL)
	if err != nil {
		level.Error(log).Log("msg", "api url is invalid", "err", err)
		return nil,nil, err
	}
	conf.SetRepositories(cfg.Repositories)

	ghExporter := exporter.Exporter{
		APIMetrics: exporter.AddMetrics(),
		Config:     conf,
	}
	r := prometheus.NewRegistry()
	err = r.Register(&ghExporter)
	if err != nil {
		level.Error(log).Log("error", err)
	}
	return &ghExporter,r, nil
}

func (rw *Github) Configure(cfg *config.Github) {
	rw.configMut.Lock()
	defer rw.configMut.Unlock()

	// Update our most recent config
	rw.cfg = cfg

	select {
	case rw.reloadCfg <- struct{}{}:
	default:
		// Something is already queued, don't need to do anything
	}
}

// Run runs the RemoteWrite until ctx is canceled. The updated function is
// unused; RemoteWrite has no observable state that can be referenced.
func (rw *Github) Run(ctx context.Context, updated func()) {
	timer1 := time.NewTimer(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-rw.reloadCfg:
			// Grab the config out of the mutex
			rw.configMut.Lock()
			conf := rw.cfg
			rw.configMut.Unlock()

			gh, err := newGithubExporter(*conf, rw.log)
			if err != nil {
				level.Error(rw.log).Log("msg", "failed to apply github config", "err", err)
			}
			rw.ex = gh
		case <-timer1.C:
			metrics, err := rw.registry.Gather()

		}
	}
}
