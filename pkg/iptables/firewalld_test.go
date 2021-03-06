package iptables

import (
	"net"
	"strconv"
	"testing"
)

func TestFirewalldInit(t *testing.T) {
	FirewalldInit()
}

func TestReloaded(t *testing.T) {
	var err error
	var fwdChain *Chain

	fwdChain, err = NewChain("FWD", "lo", Filter)
	if err != nil {
		t.Fatal(err)
	}
	defer fwdChain.Remove()

	// copy-pasted from iptables_test:TestLink
	ip1 := net.ParseIP("192.168.1.1")
	ip2 := net.ParseIP("192.168.1.2")
	port := 1234
	proto := "tcp"

	err = fwdChain.Link(Append, ip1, ip2, port, proto)
	if err != nil {
		t.Fatal(err)
	} else {
		// to be re-called again later
		OnReloaded(func() { fwdChain.Link(Append, ip1, ip2, port, proto) })
	}

	rule1 := []string{
		"-i", fwdChain.Bridge,
		"-o", fwdChain.Bridge,
		"-p", proto,
		"-s", ip1.String(),
		"-d", ip2.String(),
		"--dport", strconv.Itoa(port),
		"-j", "ACCEPT"}

	if !Exists(fwdChain.Table, fwdChain.Name, rule1...) {
		t.Fatalf("rule1 does not exist")
	}

	// flush all rules
	fwdChain.Remove()

	reloaded()

	// make sure the rules have been recreated
	if !Exists(fwdChain.Table, fwdChain.Name, rule1...) {
		t.Fatalf("rule1 hasn't been recreated")
	}
}

func TestPassthrough(t *testing.T) {
	rule1 := []string{
		"-i", "lo",
		"-p", "udp",
		"--dport", "123",
		"-j", "ACCEPT"}

	if firewalldRunning {
		_, err := Passthrough(Iptables, append([]string{"-A"}, rule1...)...)
		if err != nil {
			t.Fatal(err)
		}
		if !Exists(Filter, "INPUT", rule1...) {
			t.Fatalf("rule1 does not exist")
		}
	}

}
