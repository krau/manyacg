package common

import (
	"time"

	"github.com/imroc/req/v3"
	"github.com/krau/manyacg/collector/config"
)

var Cilent *req.Client

func init() {
	c := req.C().ImpersonateChrome()
	c.SetCommonRetryCount(2)
	c.TLSHandshakeTimeout = time.Second * 3
	c.SetTimeout(time.Second * 5)
	c.SetMaxConnsPerHost(config.Cfg.App.MaxConcurrent)
	Cilent = c
}
