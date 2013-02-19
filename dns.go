package main

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"strings"
)

// Lookup the IP address using HypeDNS
func lookupHypeDNS(hostname string) (response string, err error) {
	c := new(dns.Client)

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(hostname), dns.TypeAAAA)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, "[fc5d:baa5:61fc:6ffd:9554:67f0:e290:7535]:53")
	if r == nil || err != nil {
		return
	}

	// Stuff must be in the answer section
	for _, a := range r.Answer {
		columns := strings.Fields(a.String()) // column 4 holds the ip address
		return padIPv6(net.ParseIP(columns[4])), nil
	}
	return
}

// Lookup the hostname for an IP address using HypeDNS
// This probably needs work but I don't really know what I'm doing :)
func reverseHypeDNSLookup(ip string) (response string, err error) {
	c := new(dns.Client)

	m := new(dns.Msg)
	thing, err := dns.ReverseAddr(ip)
	if err != nil {
		return
	}
	m.SetQuestion(thing, dns.TypePTR)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, "[fc5d:baa5:61fc:6ffd:9554:67f0:e290:7535]:53")
	if r == nil || err != nil {
		return
	}

	// Stuff must be in the answer section
	for _, a := range r.Answer {
		columns := strings.Fields(a.String()) // column 4 holds the ip address
		return columns[4], nil
	}
	return
}

// Resolve an IP to a domain name using the system DNS settings first, then HypeDNS
func resolveIP(ip string) (hostname string, err error) {
	var try2 string

	// try the system DNS setup
	result, _ := net.LookupAddr(ip)
	if len(result) > 0 {
		goto end
	}

	// Try HypeDNS
	try2, err = reverseHypeDNSLookup(ip)
	if try2 == "" || err != nil {
		err = fmt.Errorf("Unable to resolve IP address. This is usually caused by not having a route to hypedns. Please try again in a few seconds.")
		return
	}
	result = append(result, try2)
end:
	for _, addr := range result {
		hostname = addr
	}

	// Trim the trailing period becuase it annoys me
	if hostname[len(hostname)-1] == '.' {
		hostname = hostname[:len(hostname)-1]
	}
	return
}

// Resolve a hostname to an IP address using the system DNS settings first, then HypeDNS
func resolveHost(hostname string) (ips []string, err error) {
	var ip string
	// Try the system DNS setup
	result, _ := net.LookupHost(hostname)
	if len(result) > 0 {
		goto end
	}

	// Try with hypedns
	ip, err = lookupHypeDNS(hostname)

	if ip == "" || err != nil {
		err = fmt.Errorf("Unable to resolve hostname. This is usually caused by not having a route to hypedns. Please try again in a few seconds.")
		return
	}

	result = append(result, ip)

end:
	for _, addr := range result {
		tIP := net.ParseIP(addr)
		// Only grab the cjdns IP's
		if tIP[0] == 0xfc {
			ips = append(ips, padIPv6(net.ParseIP(addr)))
		}
	}

	return
}
