package external

import (
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

func (e *External) soa(state request.Request) *dns.SOA {
	header := dns.RR_Header{Name: state.Zone, Rrtype: dns.TypeSOA, Ttl: 30, Class: dns.ClassINET}

	Mbox := e.hostmaster
	Ns := "ns.dns."
	if state.Zone[0] != '.' {
		Mbox += state.Zone
		Ns += state.Zone
	}

	soa := &dns.SOA{Hdr: header,
		Mbox:    Mbox,
		Ns:      Ns,
		Serial:  12345,
		Refresh: 7200,
		Retry:   1800,
		Expire:  86400,
		Minttl:  5,
	}
	return soa
}
