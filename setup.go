package loadbalancer

import (
	"fmt"
	"strconv"

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
	c.Next()
	if c.NextArg() || err != nil {
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
	// get remaining args before block
	// Skip plugin name.
	c.Next()
	// Set root zones
	e.RootZones = c.RemainingArgs()
	//fmt.Printf("Adding Root Zones: %s", e.RootZones)
	for c.NextBlock() {
		//fmt.Printf("next arg is %s\n", c.Val())
		switch c.Val() {
		case "enableNSZone":
			//fmt.Printf("Processing enableNSZone")
			if !c.NextArg() {
				return &LoadBalancer{}, c.Errf("%s missing argument", "enableNSZone")
			}
			value, err := strconv.ParseBool(c.Val())
			e.EnableNSZone = value
			fmt.Printf("EnableNSZone equals %t", e.EnableNSZone)
			if err != nil {
				return &LoadBalancer{}, c.Errf("%s unable to parse.. got %s", "enableNSZone", err)
			}
			//fmt.Print(c.Val())
		case "enableRootZone":
			fmt.Printf("Processing enableRootZone")
			if !c.NextArg() {
				return &LoadBalancer{}, c.Errf("%s missing argument", "enableRootZone")
			}
			value, err := strconv.ParseBool(c.Val())
			e.EnableRootZone = value
			if err != nil {
				return &LoadBalancer{}, c.Errf("%s unable to parse.. got %s", "enableRootZone", err)
			}
			fmt.Print(c.Val())
		case "kubeConfigPath":
			//fmt.Printf("Processing kubeConfigPath")
			if !c.NextArg() {
				return &LoadBalancer{}, c.Errf("%s missing argument", "enableRootZone")
			}
			e.KubeConfigPath = c.Val()
		default:
			if c.Val() != "}" {
				return &LoadBalancer{}, c.Errf("unknown property '%s'", c.Val())
			}
		}
	}
	return e, nil
}
