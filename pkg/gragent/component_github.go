package gragent

import (
	"agentflow/pkg/components"
	"agentflow/pkg/config"
	"context"
	"github.com/go-kit/log"
	"github.com/hashicorp/hcl/v2"
	config2 "github.com/prometheus/prometheus/config"
)

type githubComponent struct {
	id  string
	gh  *components.Github
	cfg config.Github
	log log.Logger
}

func newGithubComponent(id string, logger log.Logger) *githubComponent {
	return &githubComponent{
		id:  id,
		log: logger,
	}
}

func (c *githubComponent) Name() string { return c.id }

func (c *githubComponent) Evaluate(ectx *hcl.EvalContext, b hcl.Body) (interface{}, hcl.Diagnostics) {
	var cfg config.Github

	diags := config.DecodeHCL(ectx, b, &cfg)
	if diags.HasErrors() {
		return nil, diags
	}
	c.cfg = cfg

	return cfg, diags
}

func (c *githubComponent) CurrentState() interface{} {
	// There's no exposed state from remoteWriteComponent
	return nil
}

func (c *githubComponent) Run(ctx context.Context, onStateChange func()) {
	github, err := components.NewGithub(c.cfg, c.log)
	c.gh = github
	if err != nil {
		return
	}
	def := config2.DefaultRemoteWriteConfig
	def.Name = c.id
	c.gh.
		c.rw.Configure(&def)
	c.rw.Run(ctx, onStateChange)
}
