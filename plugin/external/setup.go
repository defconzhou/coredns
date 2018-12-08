package external

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("external", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	e, err := parse(c)
	if err != nil {
		return plugin.Error("external", err)
	}

	// Do this in OnStartup, so all plugins have been initialized.
	c.OnStartup(func() error {
		m := dnsserver.GetConfig(c).Handler("kubernetes")
		if m == nil {
			return nil
		}
		if x, ok := m.(Externaler); ok {
			e.externalFunc = x.External
		}
		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		e.Next = next
		return e
	})

	return nil
}

func parse(c *caddy.Controller) (*External, error) {
	e := &External{hostmaster: "hostmaster."} // not closing dot.

	for c.Next() {
		zones := c.RemainingArgs()
		e.Zones = zones
		if len(zones) == 0 {
			e.Zones = make([]string, len(c.ServerBlockKeys))
			copy(e.Zones, c.ServerBlockKeys)
		}
		for i, str := range e.Zones {
			e.Zones[i] = plugin.Host(str).Normalize()
		}
	}
	return e, nil
}
