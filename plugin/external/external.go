/*
Package external implements external names for kubernetes clusters.

This plugin only handles three qtypes (except the apex queries, because those are handled
differently). We support A, AAAA and SRV request, for all other types we return NODATA or
NXDOMAIN depending on the state of the cluster.
*/
package external

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

// Func defines the function a plugin should implement to return set of services. This function will never
// be called with the apex of a zone.
type Func func(request.Request) ([]msg.Service, int)

// Externaler defines the interface that a plugin should implement in order to be used by External.
type Externaler interface {
	External(request.Request) ([]msg.Service, int)
}

// External resolves Ingress and Loadbalance IPs from kubernetes clusters
type External struct {
	Next  plugin.Handler
	Zones []string

	hostmaster string

	externalFunc Func
}

// ServeDNS implements the plugin.Handle interface.
func (e *External) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	zone := plugin.Zones(e.Zones).Matches(state.Name())
	if zone == "" {
		return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
	}

	if e.externalFunc == nil {
		return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
	}

	state.Zone = zone

	// TODO: apex handling, NS handling and whatever.
	// We can have options in the plugin to tweak this.

	svc, rcode := e.externalFunc(state)

	m := new(dns.Msg)
	m.SetReply(state.Req)

	if len(svc) == 0 {
		m.Rcode = rcode
		m.Ns = []dns.RR{e.soa(state)}
		w.WriteMsg(m)
		return 0, nil
	}

	switch state.QType() {
	case dns.TypeA:
		m.Answer = a(svc, state)
	case dns.TypeAAAA:
		m.Answer = aaaa(svc, state)
	case dns.TypeSRV:
		m.Answer, m.Extra = srv(svc, state)
	default:
		m.Ns = []dns.RR{e.soa(state)}
	}

	// If we did had records, but queries for the wrong type return a nodata response.
	if len(m.Answer) == 0 {
		m.Ns = []dns.RR{e.soa(state)}
	}

	w.WriteMsg(m)
	return 0, nil
}

// Name implements the Handler interface.
func (e *External) Name() string { return "external" }
