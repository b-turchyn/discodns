package main

import (
	"testing"

	"github.com/miekg/dns"
)

func TestFilters(t *testing.T) {
	// Enable debug logging
	logDebug = true
}

func TestNoFilters(t *testing.T) {
	filterer := QueryFilterer{}
	msg := generateDNSMessage("discodns.net", dns.TypeA)

	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}
}

func TestSimpleAccept(t *testing.T) {
	filterer := QueryFilterer{acceptFilters: parseFilters([]string{"net:A"})}

	msg := generateDNSMessage("discodns.net", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeAAAA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}

	msg = generateDNSMessage("discodns.com", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}
}

func TestSimpleReject(t *testing.T) {
	filterer := QueryFilterer{rejectFilters: parseFilters([]string{"net:A"})}

	msg := generateDNSMessage("discodns.com", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeAAAA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}
}

func TestSimpleAcceptFullDomain(t *testing.T) {
	filterer := QueryFilterer{acceptFilters: parseFilters([]string{"net:"})}

	msg := generateDNSMessage("discodns.net", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeANY)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.com", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}

	msg = generateDNSMessage("discodns.com", dns.TypeANY)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}
}

func TestSimpleRejectFullDomain(t *testing.T) {
	filterer := QueryFilterer{rejectFilters: parseFilters([]string{"net:"})}

	msg := generateDNSMessage("discodns.net", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeANY)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}

	msg = generateDNSMessage("discodns.com", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.com", dns.TypeANY)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}
}

func TestSimpleAcceptSpecificTypes(t *testing.T) {
	filterer := QueryFilterer{acceptFilters: parseFilters([]string{":A"})}

	msg := generateDNSMessage("discodns.net", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeAAAA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}
}

func TestSimpleAcceptMultipleTypes(t *testing.T) {
	filterer := QueryFilterer{acceptFilters: parseFilters([]string{":A,PTR"})}

	msg := generateDNSMessage("discodns.net", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeAAAA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}

	msg = generateDNSMessage("discodns.net", dns.TypePTR)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}
}

func TestSimpleRejectSpecificTypes(t *testing.T) {
	filterer := QueryFilterer{rejectFilters: parseFilters([]string{":A"})}

	msg := generateDNSMessage("discodns.net", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeAAAA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}
}

func TestSimpleRejectMultipleTypes(t *testing.T) {
	filterer := QueryFilterer{rejectFilters: parseFilters([]string{":A,PTR"})}

	msg := generateDNSMessage("discodns.net", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeAAAA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.net", dns.TypePTR)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}
}

func TestMultipleAccept(t *testing.T) {
	filterer := QueryFilterer{acceptFilters: parseFilters([]string{"net:A", "com:AAAA"})}

	msg := generateDNSMessage("discodns.net", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeAAAA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}

	msg = generateDNSMessage("discodns.com", dns.TypeAAAA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.com", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}
}

func TestMultipleReject(t *testing.T) {
	filterer := QueryFilterer{rejectFilters: parseFilters([]string{"net:A", "com:AAAA"})}

	msg := generateDNSMessage("discodns.net", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}

	msg = generateDNSMessage("discodns.net", dns.TypeAAAA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}

	msg = generateDNSMessage("discodns.com", dns.TypeAAAA)
	if filterer.ShouldAcceptQuery(msg) != false {
		t.Fatal("Expected the query to be rejected")
	}

	msg = generateDNSMessage("discodns.com", dns.TypeA)
	if filterer.ShouldAcceptQuery(msg) != true {
		t.Fatal("Expected the query to be accepted")
	}
}

// generateDNSMessage returns a simple DNS query with a single question,
// comprised of the domain and rrType given.
func generateDNSMessage(domain string, rrType uint16) *dns.Msg {
	domain = dns.Fqdn(domain)
	msg := dns.Msg{Question: []dns.Question{dns.Question{Name: domain, Qtype: rrType}}}
	return &msg
}
