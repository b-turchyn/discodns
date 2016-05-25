package main

import (
	"strings"
	"testing"

	"github.com/coreos/go-etcd/etcd"
	"github.com/miekg/dns"
)

var (
	client   = etcd.NewClient([]string{"http://127.0.0.1:4001"})
	resolver = &Resolver{etcd: client}
)

func TestEtcd(t *testing.T) {
	// Enable debug logging
	logDebug = true

	if !client.SyncCluster() {
		t.Fatal("Failed to sync etcd cluster")
	}
}

func TestGetFromStorageSingleKey(t *testing.T) {
	resolver.etcdPrefix = "TestGetFromStorageSingleKey/"
	client.Set("TestGetFromStorageSingleKey/net/disco/.A", "1.1.1.1", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	nodes, err := resolver.GetFromStorage("net/disco/.A")
	if err != nil {
		t.Fatal("Error returned from etcd", err)
	}

	if len(nodes) != 1 {
		t.Fatal("Number of nodes should be 1: ", len(nodes))
	}

	node := nodes[0]
	if node.node.Value != "1.1.1.1" {
		t.Fatal("Node value should be 1.1.1.1: ", node)
	}
}

func TestGetFromStorageNestedKeys(t *testing.T) {
	resolver.etcdPrefix = "TestGetFromStorageNestedKeys/"
	client.Set("TestGetFromStorageNestedKeys/net/disco/.A/0", "1.1.1.1", 0)
	client.Set("TestGetFromStorageNestedKeys/net/disco/.A/1", "1.1.1.2", 0)
	client.Set("TestGetFromStorageNestedKeys/net/disco/.A/2/0", "1.1.1.3", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	nodes, err := resolver.GetFromStorage("net/disco/.A")
	if err != nil {
		t.Fatal("Error returned from etcd", err)
	}

	if len(nodes) != 3 {
		t.Fatal("Number of nodes should be 3: ", len(nodes))
	}

	var node *EtcdRecord

	node = nodes[0]
	if node.node.Value != "1.1.1.1" {
		t.Fatal("Node value should be 1.1.1.1: ", node)
	}
	node = nodes[1]
	if node.node.Value != "1.1.1.2" {
		t.Fatal("Node value should be 1.1.1.2: ", node)
	}
	node = nodes[2]
	if node.node.Value != "1.1.1.3" {
		t.Fatal("Node value should be 1.1.1.3: ", node)
	}
}

func TestNameToKeyConverter(t *testing.T) {
	var key string

	key = nameToKey("foo.net.", "")
	if key != "/net/foo" {
		t.Error("Expected key /net/foo")
	}

	key = nameToKey("foo.net", "")
	if key != "/net/foo" {
		t.Error("Expected key /net/foo")
	}

	key = nameToKey("foo.net.", "/.A")
	if key != "/net/foo/.A" {
		t.Error("Expected key /net/foo/.A")
	}
}

/**
 * Test that the right authority is being returned for different types of DNS
 * queries.
 */

func TestAuthorityRoot(t *testing.T) {
	resolver.etcdPrefix = "TestAuthorityRoot/"
	client.Set("TestAuthorityRoot/net/disco/.SOA", "ns1.disco.net.\tadmin.disco.net.\t3600\t600\t86400\t10", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	query := new(dns.Msg)
	query.SetQuestion("disco.net.", dns.TypeA)

	answer := resolver.Lookup(query)

	if len(answer.Answer) > 0 {
		t.Fatal("Expected zero answers")
	}

	if len(answer.Ns) != 1 {
		t.Fatal("Expected one authority record")
	}

	rr := answer.Ns[0].(*dns.SOA)
	header := rr.Header()

	// Verify the header is correct
	if header.Name != "disco.net." {
		t.Fatal("Expected record with name disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeSOA {
		t.Fatal("Expected record with type SOA:", header.Rrtype)
	}

	// Verify the record itself is correct
	if rr.Ns != "ns1.disco.net." {
		t.Fatal("Expected NS to be ns1.disco.net.: ", rr.Ns)
	}
	if rr.Mbox != "admin.disco.net." {
		t.Fatal("Expected MBOX to be admin.disco.net.: ", rr.Mbox)
	}
	// if rr.Serial != "admin.disco.net" {
	//     t.Error("Expected MBOX to be admin.disco.net: ", rr.Mbox)
	// }
	if rr.Refresh != 3600 {
		t.Fatal("Expected REFRESH to be 3600: ", rr.Refresh)
	}
	if rr.Retry != 600 {
		t.Fatal("Expected RETRY to be 600: ", rr.Retry)
	}
	if rr.Expire != 86400 {
		t.Fatal("Expected EXPIRE to be 86400: ", rr.Expire)
	}
	if rr.Minttl != 10 {
		t.Fatal("Expected MINTTL to be 10: ", rr.Minttl)
	}
}

func TestAuthorityDomain(t *testing.T) {
	resolver.etcdPrefix = "TestAuthorityDomain/"
	client.Set("TestAuthorityDomain/net/disco/.SOA", "ns1.disco.net.\tadmin.disco.net.\t3600\t600\t86400\t10", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	query := new(dns.Msg)
	query.SetQuestion("bar.disco.net.", dns.TypeA)

	answer := resolver.Lookup(query)

	if len(answer.Answer) > 0 {
		t.Fatal("Expected zero answers")
	}

	if len(answer.Ns) != 1 {
		t.Fatal("Expected one authority record")
	}

	rr := answer.Ns[0].(*dns.SOA)
	header := rr.Header()

	// Verify the header is correct
	if header.Name != "disco.net." {
		t.Fatal("Expected record with name disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeSOA {
		t.Fatal("Expected record with type SOA:", header.Rrtype)
	}

	// Verify the record itself is correct
	if rr.Ns != "ns1.disco.net." {
		t.Fatal("Expected NS to be ns1.disco.net.: ", rr.Ns)
	}
	if rr.Mbox != "admin.disco.net." {
		t.Fatal("Expected MBOX to be admin.disco.net.: ", rr.Mbox)
	}
	if rr.Refresh != 3600 {
		t.Fatal("Expected REFRESH to be 3600: ", rr.Refresh)
	}
	if rr.Retry != 600 {
		t.Fatal("Expected RETRY to be 600: ", rr.Retry)
	}
	if rr.Expire != 86400 {
		t.Fatal("Expected EXPIRE to be 86400: ", rr.Expire)
	}
	if rr.Minttl != 10 {
		t.Fatal("Expected MINTTL to be 10: ", rr.Minttl)
	}
}

func TestAuthoritySubdomain(t *testing.T) {
	resolver.etcdPrefix = "TestAuthoritySubdomain/"
	client.Set("TestAuthoritySubdomain/net/disco/.SOA", "ns1.disco.net.\tadmin.disco.net.\t3600\t600\t86400\t10", 0)
	client.Set("TestAuthoritySubdomain/net/disco/bar/.SOA", "ns1.bar.disco.net.\tbar.disco.net.\t3600\t600\t86400\t10", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	query := new(dns.Msg)
	query.SetQuestion("foo.bar.disco.net.", dns.TypeA)

	answer := resolver.Lookup(query)

	if len(answer.Answer) > 0 {
		t.Fatal("Expected zero answers")
	}

	if len(answer.Ns) != 1 {
		t.Fatal("Expected one authority record")
	}

	rr := answer.Ns[0].(*dns.SOA)
	header := rr.Header()

	// Verify the header is correct
	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeSOA {
		t.Fatal("Expected record with type SOA:", header.Rrtype)
	}

	// Verify the record itself is correct
	if rr.Ns != "ns1.bar.disco.net." {
		t.Fatal("Expected NS to be ns1.disco.net.: ", rr.Ns)
	}
	if rr.Mbox != "bar.disco.net." {
		t.Fatal("Expected MBOX to be admin.disco.net.: ", rr.Mbox)
	}
	if rr.Refresh != 3600 {
		t.Fatal("Expected REFRESH to be 3600: ", rr.Refresh)
	}
	if rr.Retry != 600 {
		t.Fatal("Expected RETRY to be 600: ", rr.Retry)
	}
	if rr.Expire != 86400 {
		t.Fatal("Expected EXPIRE to be 86400: ", rr.Expire)
	}
	if rr.Minttl != 10 {
		t.Fatal("Expected MINTTL to be 10: ", rr.Minttl)
	}
}

/**
 * Test different that types of DNS queries return the correct answers
 **/

func TestAnswerQuestionA(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionA/"
	client.Set("TestAnswerQuestionA/net/disco/bar/.A", "1.2.3.4", 0)
	client.Set("TestAnswerQuestionA/net/disco/.SOA", "ns1.disco.net.\tadmin.disco.net.\t3600\t600\t86400\t10", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	query := new(dns.Msg)
	query.SetQuestion("bar.disco.net.", dns.TypeA)

	answer := resolver.Lookup(query)

	if len(answer.Answer) != 1 {
		t.Fatal("Expected one answer, got ", len(answer.Answer))
	}

	if len(answer.Ns) > 0 {
		t.Fatal("Didn't expect any authority records")
	}

	rr := answer.Answer[0].(*dns.A)
	header := rr.Header()

	// Verify the header is correct
	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeA {
		t.Fatal("Expected record with type A:", header.Rrtype)
	}

	// Verify the record itself is correct
	if rr.A.String() != "1.2.3.4" {
		t.Fatal("Expected A record to be 1.2.3.4: ", rr.A)
	}
}

func TestAnswerQuestionAAAA(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionAAAA/"
	client.Set("TestAnswerQuestionAAAA/net/disco/bar/.AAAA", "::1", 0)
	client.Set("TestAnswerQuestionAAAA/net/disco/.SOA", "ns1.disco.net.\tadmin.disco.net.\t3600\t600\t86400\t10", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	query := new(dns.Msg)
	query.SetQuestion("bar.disco.net.", dns.TypeAAAA)

	answer := resolver.Lookup(query)

	if len(answer.Answer) != 1 {
		t.Fatal("Expected one answer, got ", len(answer.Answer))
	}

	if len(answer.Ns) > 0 {
		t.Fatal("Didn't expect any authority records")
	}

	rr := answer.Answer[0].(*dns.AAAA)
	header := rr.Header()

	// Verify the header is correct
	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeAAAA {
		t.Fatal("Expected record with type AAAA:", header.Rrtype)
	}

	// Verify the record itself is correct
	if rr.AAAA.String() != "::1" {
		t.Fatal("Expected AAAA record to be ::1: ", rr.AAAA)
	}
}

func TestAnswerQuestionANY(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionANY/"
	client.Set("TestAnswerQuestionANY/net/disco/bar/.TXT", "google.com.", 0)
	client.Set("TestAnswerQuestionANY/net/disco/bar/.A/0", "1.2.3.4", 0)
	client.Set("TestAnswerQuestionANY/net/disco/bar/.A/1", "2.3.4.5", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	query := new(dns.Msg)
	query.SetQuestion("bar.disco.net.", dns.TypeANY)

	answer := resolver.Lookup(query)

	if len(answer.Answer) != 3 {
		t.Fatal("Expected one answer, got ", len(answer.Answer))
	}

	if len(answer.Ns) > 0 {
		t.Fatal("Didn't expect any authority records")
	}
}

func TestAnswerQuestionUnsupportedType(t *testing.T) {
	// query for a type that we don't have support for (I tried to pick the most
	// obscure rr type that the dns library supports and that we're unlikely to
	// add support for)
	query := new(dns.Msg)
	query.SetQuestion("bar.disco.net.", dns.TypeEUI64)

	answer := resolver.Lookup(query)

	if len(answer.Answer) != 0 {
		t.Fatal("Expected no answers, got ", len(answer.Answer))
	}

	if answer.Rcode != dns.RcodeNameError {
		t.Fatal("Expected NXDOMAIN response code, got", dns.RcodeToString[answer.Rcode])
	}

	if len(answer.Ns) > 0 {
		t.Fatal("Didn't expect any authority records")
	}
}

func TestAnswerQuestionWildcardCNAME(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionCNAME/"
	client.Set("TestAnswerQuestionCNAME/net/disco/*/.CNAME", "baz.disco.net.", 0)
	client.Set("TestAnswerQuestionCNAME/net/disco/baz/.A", "1.2.3.4", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	query := new(dns.Msg)
	query.SetQuestion("test.disco.net.", dns.TypeA)

	answer := resolver.Lookup(query)

	if len(answer.Answer) != 1 {
		t.Fatal("Expected one answers, got ", len(answer.Answer))
	}

	if len(answer.Ns) > 0 {
		t.Fatal("Didn't expect any authority records")
	}

	rr := answer.Answer[0].(*dns.CNAME)
	header := rr.Header()

	// Verify the header is correct
	if header.Name != "test.disco.net." {
		t.Fatal("Expected record with name test.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeCNAME {
		t.Fatal("Expected record with type AAAA:", header.Rrtype)
	}

	// Verify the CNAME data is correct
	if rr.Target != "baz.disco.net." {
		t.Fatal("Expected CNAME target baz.disco.net.:", header.Rrtype)
	}
}

func TestAnswerQuestionCNAME(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionCNAME/"
	client.Set("TestAnswerQuestionCNAME/net/disco/bar/.CNAME", "baz.disco.net.", 0)
	client.Set("TestAnswerQuestionCNAME/net/disco/baz/.A", "1.2.3.4", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	query := new(dns.Msg)
	query.SetQuestion("bar.disco.net.", dns.TypeA)

	answer := resolver.Lookup(query)

	if len(answer.Answer) != 1 {
		t.Fatal("Expected one answers, got ", len(answer.Answer))
	}

	if len(answer.Ns) > 0 {
		t.Fatal("Didn't expect any authority records")
	}

	rr := answer.Answer[0].(*dns.CNAME)
	header := rr.Header()

	// Verify the header is correct
	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeCNAME {
		t.Fatal("Expected record with type CNAME:", header.Rrtype)
	}

	// Verify the CNAME data is correct
	if rr.Target != "baz.disco.net." {
		t.Fatal("Expected CNAME target baz.disco.net.:", header.Rrtype)
	}
}

func TestAnswerQuestionWildcardAAAANoMatch(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionWildcardANoMatch/"
	client.Set("TestAnswerQuestionWildcardANoMatch/net/disco/bar/*/.AAAA", "::1", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	query := new(dns.Msg)
	query.SetQuestion("bar.disco.net.", dns.TypeAAAA)

	answer := resolver.Lookup(query)

	if len(answer.Answer) > 0 {
		t.Fatal("Didn't expect any answers, got ", len(answer.Answer))
	}
}

func TestAnswerQuestionWildcardAAAA(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionWildcardA/"
	client.Set("TestAnswerQuestionWildcardA/net/disco/bar/*/.AAAA", "::1", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	query := new(dns.Msg)
	query.SetQuestion("baz.bar.disco.net.", dns.TypeAAAA)

	answer := resolver.Lookup(query)

	if len(answer.Answer) != 1 {
		t.Fatal("Expected one answer, got ", len(answer.Answer))
	}

	if len(answer.Ns) > 0 {
		t.Fatal("Didn't expect any authority records")
	}

	rr := answer.Answer[0].(*dns.AAAA)
	header := rr.Header()

	// Verify the header is correct
	if header.Name != "baz.bar.disco.net." {
		t.Fatal("Expected record with name baz.bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeAAAA {
		t.Fatal("Expected record with type AAAA:", header.Rrtype)
	}

	// Verify the record itself is correct
	if rr.AAAA.String() != "::1" {
		t.Fatal("Expected AAAA record to be ::1: ", rr.AAAA)
	}
}

func TestAnswerQuestionTTL(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionTTL/"
	client.Set("TestAnswerQuestionTTL/net/disco/bar/.A", "1.2.3.4", 0)
	client.Set("TestAnswerQuestionTTL/net/disco/bar/.A.ttl", "300", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeA)

	if len(records) != 1 {
		t.Fatal("Expected one answer, got ", len(records))
	}

	rr := records[0].(*dns.A)
	header := rr.Header()

	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeA {
		t.Fatal("Expected record with type A:", header.Rrtype)
	}
	if header.Ttl != 300 {
		t.Fatal("Expected TTL of 300 seconds:", header.Ttl)
	}
	if rr.A.String() != "1.2.3.4" {
		t.Fatal("Expected A record to be 1.2.3.4: ", rr.A)
	}
}

func TestAnswerQuestionTTLMultipleRecords(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionTTLMultipleRecords/"
	client.Set("TestAnswerQuestionTTLMultipleRecords/net/disco/bar/.A/0", "1.2.3.4", 0)
	client.Set("TestAnswerQuestionTTLMultipleRecords/net/disco/bar/.A/0.ttl", "300", 0)
	client.Set("TestAnswerQuestionTTLMultipleRecords/net/disco/bar/.A/1", "8.8.8.8", 0)
	client.Set("TestAnswerQuestionTTLMultipleRecords/net/disco/bar/.A/1.ttl", "600", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeA)

	if len(records) != 2 {
		t.Fatal("Expected two answers, got ", len(records))
	}

	rrOne := records[0].(*dns.A)
	headerOne := rrOne.Header()

	if headerOne.Ttl != 300 {
		t.Fatal("Expected TTL of 300 seconds:", headerOne.Ttl)
	}
	if rrOne.A.String() != "1.2.3.4" {
		t.Fatal("Expected A record to be 1.2.3.4: ", rrOne.A)
	}

	rrTwo := records[1].(*dns.A)
	headerTwo := rrTwo.Header()

	if headerTwo.Ttl != 600 {
		t.Fatal("Expected TTL of 300 seconds:", headerTwo.Ttl)
	}
	if rrTwo.A.String() != "8.8.8.8" {
		t.Fatal("Expected A record to be 8.8.8.8: ", rrTwo.A)
	}
}

func TestAnswerQuestionTTLInvalidFormat(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionTTL/"
	client.Set("TestAnswerQuestionTTL/net/disco/bar/.A", "1.2.3.4", 0)
	client.Set("TestAnswerQuestionTTL/net/disco/bar/.A.ttl", "haha", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeA)

	if len(records) != 1 {
		t.Fatal("Expected one answer, got ", len(records))
	}

	rr := records[0].(*dns.A)
	header := rr.Header()

	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeA {
		t.Fatal("Expected record with type A:", header.Rrtype)
	}
	if header.Ttl != 0 {
		t.Fatal("Expected TTL of 0 seconds:", header.Ttl)
	}
	if rr.A.String() != "1.2.3.4" {
		t.Fatal("Expected A record to be 1.2.3.4: ", rr.A)
	}
}

func TestAnswerQuestionTTLDanglingNode(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionTTLDanglingNode/"
	client.Set("TestAnswerQuestionTTLDanglingNode/net/disco/bar/.TXT.ttl", "600", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeTXT)

	if len(records) != 0 {
		t.Fatal("Expected no answer, got ", len(records))
	}
}

func TestAnswerQuestionTTLDanglingDirNode(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionTTLDanglingDirNode/"
	client.Set("TestAnswerQuestionTTLDanglingDirNode/net/disco/bar/.TXT/0.ttl", "600", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeTXT)

	if len(records) != 0 {
		t.Fatal("Expected no answer, got ", len(records))
	}
}

func TestAnswerQuestionTTLDanglingDirSibling(t *testing.T) {
	resolver.etcdPrefix = "TestAnswerQuestionTTLDanglingDirSibling/"
	client.Set("TestAnswerQuestionTTLDanglingDirSibling/net/disco/bar/.TXT/0.ttl", "100", 0)
	client.Set("TestAnswerQuestionTTLDanglingDirSibling/net/disco/bar/.TXT/1", "foo bar", 0)
	client.Set("TestAnswerQuestionTTLDanglingDirSibling/net/disco/bar/.TXT/1.ttl", "600", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeTXT)

	if len(records) != 1 {
		t.Fatal("Expected one answer, got ", len(records))
	}

	rr := records[0].(*dns.TXT)
	header := rr.Header()

	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeTXT {
		t.Fatal("Expected record with type TXT:", header.Rrtype)
	}
	if header.Ttl != 600 {
		t.Fatal("Expected TTL of 600 seconds:", header.Ttl)
	}
	if strings.Join(rr.Txt, "\n") != "foo bar" {
		t.Fatal("Expected txt record to be 'foo bar': ", rr.Txt)
	}
}

/**
 * Test converstion of names (i.e etcd nodes) to single records of different
 * types.
 **/

func TestLookupAnswerForA(t *testing.T) {
	resolver.etcdPrefix = "TestLookupAnswerForA/"
	client.Set("TestLookupAnswerForA/net/disco/bar/.A", "1.2.3.4", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeA)

	if len(records) != 1 {
		t.Fatal("Expected one answer, got ", len(records))
	}

	rr := records[0].(*dns.A)
	header := rr.Header()

	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeA {
		t.Fatal("Expected record with type A:", header.Rrtype)
	}
	if rr.A.String() != "1.2.3.4" {
		t.Fatal("Expected A record to be 1.2.3.4: ", rr.A)
	}
}

func TestLookupAnswerForAAAA(t *testing.T) {
	resolver.etcdPrefix = "TestLookupAnswerForAAAA/"
	client.Set("TestLookupAnswerForAAAA/net/disco/bar/.AAAA", "::1", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeAAAA)

	if len(records) != 1 {
		t.Fatal("Expected one answer, got ", len(records))
	}

	rr := records[0].(*dns.AAAA)
	header := rr.Header()

	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeAAAA {
		t.Fatal("Expected record with type AAAA:", header.Rrtype)
	}
	if rr.AAAA.String() != "::1" {
		t.Fatal("Expected AAAA record to be ::1: ", rr.AAAA)
	}
}

func TestLookupAnswerForCNAME(t *testing.T) {
	resolver.etcdPrefix = "TestLookupAnswerForCNAME/"
	client.Set("TestLookupAnswerForCNAME/net/disco/bar/.CNAME", "cname.google.com.", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeCNAME)

	if len(records) != 1 {
		t.Fatal("Expected one answer, got ", len(records))
	}

	rr := records[0].(*dns.CNAME)
	header := rr.Header()

	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeCNAME {
		t.Fatal("Expected record with type CNAME:", header.Rrtype)
	}
	if rr.Target != "cname.google.com." {
		t.Fatal("Expected CNAME record to be cname.google.com.: ", rr.Target)
	}
}

func TestLookupAnswerForNS(t *testing.T) {
	resolver.etcdPrefix = "TestLookupAnswerForNS/"
	client.Set("TestLookupAnswerForNS/net/disco/bar/.NS", "dns.google.com.", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeNS)

	if len(records) != 1 {
		t.Fatal("Expected one answer, got ", len(records))
	}

	rr := records[0].(*dns.NS)
	header := rr.Header()

	if header.Name != "bar.disco.net." {
		t.Fatal("Expected record with name bar.disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeNS {
		t.Fatal("Expected record with type NS:", header.Rrtype)
	}
	if rr.Ns != "dns.google.com." {
		t.Fatal("Expected NS record to be dns.google.com.: ", rr.Ns)
	}
}

func TestLookupAnswerForSOA(t *testing.T) {
	resolver.etcdPrefix = "TestLookupAnswerForSOA/"
	client.Set("TestLookupAnswerForSOA/net/disco/.SOA", "ns1.disco.net.\tadmin.disco.net.\t3600\t600\t86400\t10", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("disco.net.", dns.TypeSOA)

	if len(records) != 1 {
		t.Fatal("Expected one answer, got ", len(records))
	}

	rr := records[0].(*dns.SOA)
	header := rr.Header()

	if header.Name != "disco.net." {
		t.Fatal("Expected record with name disco.net.: ", header.Name)
	}
	if header.Rrtype != dns.TypeSOA {
		t.Fatal("Expected record with type SOA:", header.Rrtype)
	}

	// Verify the record itself is correct
	if rr.Ns != "ns1.disco.net." {
		t.Fatal("Expected NS to be ns1.disco.net.: ", rr.Ns)
	}
	if rr.Mbox != "admin.disco.net." {
		t.Fatal("Expected MBOX to be admin.disco.net.: ", rr.Mbox)
	}
	if rr.Refresh != 3600 {
		t.Fatal("Expected REFRESH to be 3600: ", rr.Refresh)
	}
	if rr.Retry != 600 {
		t.Fatal("Expected RETRY to be 600: ", rr.Retry)
	}
	if rr.Expire != 86400 {
		t.Fatal("Expected EXPIRE to be 86400: ", rr.Expire)
	}
	if rr.Minttl != 10 {
		t.Fatal("Expected MINTTL to be 10: ", rr.Minttl)
	}
}

func TestLookupAnswerForPTR(t *testing.T) {
	resolver.etcdPrefix = "TestLookupAnswerForPTR/"
	client.Set("TestLookupAnswerForPTR/net/disco/alias/.PTR/target1", "target1.disco.net.", 0)
	client.Set("TestLookupAnswerForPTR/net/disco/alias/.PTR/target2", "target2.disco.net.", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("alias.disco.net.", dns.TypePTR)

	if len(records) != 2 {
		t.Fatal("Expected two answers, got ", len(records))
	}

	seen1 := false
	seen2 := false

	// We can't (and shouldn't try to) guarantee order, so check for all
	// expected records the long way
	for _, record := range records {
		rr := record.(*dns.PTR)
		header := rr.Header()

		if header.Rrtype != dns.TypePTR {
			t.Fatal("Expected record with type PTR:", header.Rrtype)
		}

		t.Log(rr)

		if rr.Ptr == "target1.disco.net." {
			seen1 = true
		}

		if rr.Ptr == "target2.disco.net." {
			seen2 = true
		}
	}

	if seen1 == false || seen2 == false {
		t.Fatal("Didn't get back all expected PTR responses")
	}
}

func TestLookupAnswerForPTRInvalidDomain(t *testing.T) {
	resolver.etcdPrefix = "TestLookupAnswerForPTRInvalidDomain/"
	client.Set("TestLookupAnswerForPTRInvalidDomain/net/disco/bad-alias/.PTR", "...", 0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, err := resolver.LookupAnswersForType("bad-alias.disco.net.", dns.TypePTR)

	if len(records) > 0 {
		t.Fatal("Expected no answers, got ", len(records))
	}

	if err == nil {
		t.Fatal("Expected error, didn't get one")
	}
}

func TestLookupAnswerForSRV(t *testing.T) {
	resolver.etcdPrefix = "TestLookupAnswerForSRV/"
	client.Set("TestLookupAnswerForSRV/net/disco/_tcp/_http/.SRV",
		"100\t100\t80\tsome-webserver.disco.net",
		0)
	defer client.Delete(resolver.etcdPrefix, true)

	records, _ := resolver.LookupAnswersForType("_http._tcp.disco.net.", dns.TypeSRV)

	if len(records) != 1 {
		t.Fatal("Expected one answer, got ", len(records))
	}

	rr := records[0].(*dns.SRV)

	if rr.Priority != 100 {
		t.Error("Unexpected 'priority' value for SRV record:", rr.Priority)
	}

	if rr.Weight != 100 {
		t.Error("Unexpected 'weight' value for SRV record:", rr.Weight)
	}

	if rr.Port != 80 {
		t.Error("Unexpected 'port' value for SRV record:", rr.Port)
	}

	if rr.Target != "some-webserver.disco.net." {
		t.Error("Unexpected 'target' value for SRV record:", rr.Target)
	}
}

func TestLookupAnswerForSRVInvalidValues(t *testing.T) {
	resolver.etcdPrefix = "TestLookupAnswerForSRVInvalidValues/"
	defer client.Delete(resolver.etcdPrefix, true)

	var badValsMap = map[string]string{
		"wrong-delimiter":    "10 10 80 foo.disco.net",
		"not-enough-fields":  "0\t0",
		"neg-int-priority":   "-10\t10\t80\tfoo.disco.net",
		"neg-int-weight":     "10\t-10\t80\tfoo.disco.net",
		"neg-int-port":       "10\t10\t-80\tfoo.disco.net",
		"large-int-priority": "65536\t10\t80\tfoo.disco.net",
		"large-int-weight":   "10\t65536\t80\tfoo.disco.net",
		"large-int-port":     "10\t10\t65536\tfoo.disco.net"}

	for name, value := range badValsMap {

		client.Set("TestLookupAnswerForSRVInvalidValues/net/disco/"+name+"/.SRV", value, 0)
		records, err := resolver.LookupAnswersForType(name+".disco.net.", dns.TypeSRV)

		if len(records) > 0 {
			t.Fatal("Expected no answers, got ", len(records))
		}

		if err == nil {
			t.Fatal("Expected error, didn't get one")
		}
	}
}
