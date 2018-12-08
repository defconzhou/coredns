package external

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/kubernetes"
	"github.com/coredns/coredns/plugin/kubernetes/object"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/pkg/watch"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
	api "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExternal(t *testing.T) {
	k := kubernetes.New([]string{"cluster.local."})
	k.Namespaces = map[string]bool{"testns": true}
	k.APIConn = &external{}

	e := &External{}
	e.Zones = []string{"example.com."}
	e.Next = test.NextHandler(dns.RcodeSuccess, nil)
	e.externalFunc = k.External

	ctx := context.TODO()
	for i, tc := range tests {
		r := tc.Msg()
		w := dnstest.NewRecorder(&test.ResponseWriter{})

		_, err := e.ServeDNS(ctx, w, r)
		if err != tc.Error {
			t.Errorf("Test %d expected no error, got %v", i, err)
			return
		}
		if tc.Error != nil {
			continue
		}

		resp := w.Msg
		if resp == nil {
			t.Fatalf("Test %d, got nil message and no error for %q", i, r.Question[0].Name)
		}

		test.CNAMEOrder(t, resp)
		test.SortAndCheck(t, resp, tc)
	}
}

var tests = []test.Case{
	// A Service
	{
		Qname: "svc1.testns.example.com.", Qtype: dns.TypeA,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("svc1.testns.example.com.	5	IN	A	1.2.3.4"),
		},
	},
	{
		Qname: "svc1.testns.example.com.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{test.SRV("svc1.testns.example.com.	5	IN	SRV	0 100 80 svc1.testns.example.com.")},
		Extra: []dns.RR{test.A("svc1.testns.example.com.  5       IN      A       1.2.3.4")},
	},
	// SRV Service Not udp/tcp
	{
		Qname: "*._not-udp-or-tcp.svc1.testns.example.com.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeNameError,
		Ns: []dns.RR{
			test.SOA("example.com.	30	IN	SOA	ns.dns.example.com. hostmaster.example.com. 1499347823 7200 1800 86400 60"),
		},
	},
	// SRV Service
	{
		Qname: "_http._tcp.svc1.testns.example.com.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.SRV("_http._tcp.svc1.testns.example.com.	5	IN	SRV	0 100 80 svc1.testns.example.com."),
		},
		Extra: []dns.RR{
			test.A("svc1.testns.example.com.	5	IN	A	1.2.3.4"),
		},
	},
	// AAAA Service (with an existing A record, but no AAAA record)
	{
		Qname: "svc1.testns.example.com.", Qtype: dns.TypeAAAA,
		Rcode: dns.RcodeSuccess,
		Ns: []dns.RR{
			test.SOA("example.com.	30	IN	SOA	ns.dns.example.com. hostmaster.example.com. 1499347823 7200 1800 86400 60"),
		},
	},
	// AAAA Service (non-existing service)
	{
		Qname: "svc0.testns.example.com.", Qtype: dns.TypeAAAA,
		Rcode: dns.RcodeNameError,
		Ns: []dns.RR{
			test.SOA("example.com.	30	IN	SOA	ns.dns.example.com. hostmaster.example.com. 1499347823 7200 1800 86400 60"),
		},
	},
	// A Service (non-existing service)
	{
		Qname: "svc0.testns.example.com.", Qtype: dns.TypeA,
		Rcode: dns.RcodeNameError,
		Ns: []dns.RR{
			test.SOA("example.com.	30	IN	SOA	ns.dns.example.com. hostmaster.example.com. 1499347823 7200 1800 86400 60"),
		},
	},
	// A Service (non-existing namespace)
	{
		Qname: "svc0.svc-nons.example.com.", Qtype: dns.TypeA,
		Rcode: dns.RcodeNameError,
		Ns: []dns.RR{
			test.SOA("example.com.	30	IN	SOA	ns.dns.example.com. hostmaster.example.com. 1499347823 7200 1800 86400 60"),
		},
	},
	// AAAA Service
	{
		Qname: "svc6.testns.example.com.", Qtype: dns.TypeAAAA,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.AAAA("svc6.testns.example.com.	5	IN	AAAA	1:2::5"),
		},
	},
	// SRV
	{
		Qname: "_http._tcp.svc6.testns.example.com.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.SRV("_http._tcp.svc6.testns.example.com.	5	IN	SRV	0 100 80 svc6.testns.example.com."),
		},
		Extra: []dns.RR{
			test.AAAA("svc6.testns.example.com.	5	IN	AAAA	1:2::5"),
		},
	},
	// SRV
	{
		Qname: "svc6.testns.example.com.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.SRV("svc6.testns.example.com.	5	IN	SRV	0 100 80 svc6.testns.example.com."),
		},
		Extra: []dns.RR{
			test.AAAA("svc6.testns.example.com.	5	IN	AAAA	1:2::5"),
		},
	},
	{
		Qname: "testns.example.com.", Qtype: dns.TypeA,
		Rcode: dns.RcodeSuccess,
		Ns: []dns.RR{
			test.SOA("example.com.	303	IN	SOA	ns.dns.example.com. hostmaster.example.com. 1499347823 7200 1800 86400 60"),
		},
	},
}

type external struct{}

func (external) HasSynced() bool                              { return true }
func (external) Run()                                         { return }
func (external) Stop() error                                  { return nil }
func (external) EpIndexReverse(string) []*object.Endpoints    { return nil }
func (external) SvcIndexReverse(string) []*object.Service     { return nil }
func (external) Modified() int64                              { return 0 }
func (external) SetWatchChan(watch.Chan)                      {}
func (external) Watch(string) error                           { return nil }
func (external) StopWatching(string)                          {}
func (external) EpIndex(s string) []*object.Endpoints         { return nil }
func (external) EndpointsList() []*object.Endpoints           { return nil }
func (external) GetNodeByName(name string) (*api.Node, error) { return nil, nil }
func (external) SvcIndex(s string) []*object.Service          { return svcIndexExternal[s] }
func (external) PodIndex(string) []*object.Pod                { return nil }

func (external) GetNamespaceByName(name string) (*api.Namespace, error) {
	return &api.Namespace{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
	}, nil
}

var svcIndexExternal = map[string][]*object.Service{
	"svc1.testns": {
		{
			Name:        "svc1",
			Namespace:   "testns",
			Type:        api.ServiceTypeClusterIP,
			ClusterIP:   "10.0.0.1",
			ExternalIPs: []string{"1.2.3.4"},
			Ports:       []api.ServicePort{{Name: "http", Protocol: "tcp", Port: 80}},
		},
	},
	"svc6.testns": {
		{
			Name:        "svc6",
			Namespace:   "testns",
			Type:        api.ServiceTypeClusterIP,
			ClusterIP:   "10.0.0.3",
			ExternalIPs: []string{"1:2::5"},
			Ports:       []api.ServicePort{{Name: "http", Protocol: "tcp", Port: 80}},
		},
	},
}

func (external) ServiceList() []*object.Service {
	var svcs []*object.Service
	for _, svc := range svcIndexExternal {
		svcs = append(svcs, svc...)
	}
	return svcs
}
