package loadbalancer

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

// init registers this plugin.
func init() { plugin.Register("k8s_loadbalancer", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {
	e, err := parse(c)
	c.Next() // Ignore "example" and give us the next token.
	if c.NextArg() || err != nil {
		// If there was another token, return an error, because we don't have any configuration.
		// Any errors returned from this setup function should be wrapped with plugin.Error, so we
		// can present a slightly nicer error message to the user.
		return plugin.Error("k8s_loadbalancer", c.ArgErr())
	}

	c.OnStartup(func() error {
		go e.updateTicker()
		return nil
	})
	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		e.Next = next
		return e
	})

	// All OK, return a nil error.
	return nil
}

func parse(c *caddy.Controller) (*LoadBalancer, error) {
	e := NewLoadBalancer()
	return e, nil
}
