package gragent

import (
	"agentflow/pkg/components"
	"agentflow/pkg/config"
	"context"
	"github.com/go-kit/log"
	cmnconfig "github.com/prometheus/common/config"
	config2 "github.com/prometheus/prometheus/config"
	"net/url"

	"github.com/hashicorp/hcl/v2"
)

type remoteWriteBlock struct {
	Name string `hcl:"name,label"`

	Body   hcl.Body `hcl:",body"`
	Remain hcl.Body `hcl:",remain"`
}

type remoteWriteComponent struct {
	id  string
	rw  *components.RemoteWrite
	cfg config.RemoteWrite
	log log.Logger
}

func newRemoteWriteComponent(id string, logger log.Logger) *remoteWriteComponent {
	return &remoteWriteComponent{
		id:  id,
		log: logger,
	}
}

func (c *remoteWriteComponent) Name() string { return c.id }

func (c *remoteWriteComponent) Evaluate(ectx *hcl.EvalContext, b hcl.Body) (interface{}, hcl.Diagnostics) {
	var cfg config.RemoteWrite

	diags := config.DecodeHCL(ectx, b, &cfg)
	if diags.HasErrors() {
		return nil, diags
	}
	c.cfg = cfg

	return cfg, diags
}

func (c *remoteWriteComponent) CurrentState() interface{} {
	// There's no exposed state from remoteWriteComponent
	return nil
}

func (c *remoteWriteComponent) Run(ctx context.Context, onStateChange func()) {
	c.rw = components.NewRemoteWrite(c.log, c.id, c.cfg.WalDir)
	def := config2.DefaultRemoteWriteConfig
	def.Name = c.id
	rwUrl, _ := url.Parse(c.cfg.URL)
	cmUrl := cmnconfig.URL{}
	cmUrl.URL = rwUrl
	def.URL = &cmUrl
	httpCfg := cmnconfig.DefaultHTTPClientConfig
	secr := cmnconfig.Secret(c.cfg.Password)
	httpCfg.BasicAuth = &cmnconfig.BasicAuth{
		Username: c.cfg.Username,
		Password: secr,
	}
	def.QueueConfig = config2.DefaultQueueConfig
	def.MetadataConfig = config2.DefaultMetadataConfig
	c.rw.Configure(&def)
	c.rw.Run(ctx, onStateChange)
}
