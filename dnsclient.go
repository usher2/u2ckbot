package main

import (
	"fmt"
	"github.com/miekg/dns"
	"math/rand"
	"strings"
	"time"
)

const ATTEMPTS = 1
const TIMEOUT = 5

func appendIfMissing(slice []string, s string) []string {
	for _, ele := range slice {
		if ele == s {
			return slice
		}
	}
	return append(slice, s)
}

func getIP4(domain string) (res []string) {
	if r, _, err := GetRR(domain, []string{"127.0.0.11:53"}, dns.TypeA); err == nil {
		switch r.Rcode {
		case dns.RcodeSuccess:
			if len(r.Answer) > 0 {
				if len(r.Answer) > 99 {
					Warning.Printf("Internal error, for %s, Answer too big: %d\n", domain, len(r.Answer))
				}
				cnt := 0
				for _, rr := range r.Answer {
					switch rr.Header().Rrtype {
					case dns.TypeA:
						if cnt > 8 {
							continue
						}
						res = appendIfMissing(res, rr.(*dns.A).A.String())
						cnt++
					case dns.TypeCNAME:
						res1 := getIP4(strings.TrimSuffix(rr.(*dns.CNAME).Target, "."))
						for _, ele := range res1 {
							res = appendIfMissing(res, ele)
						}
					case dns.TypeRRSIG:
						Debug.Printf("Warning: RRSIG (%s): %#v", domain, r.MsgHdr)
					default:
						Debug.Printf("Warning: unknown answer (%s): %s\n", domain, rr.String())
					}
				}
			}
		}
	} else {
		Error.Printf("Type A. Internal error (%s): %s\n", domain, err.Error())
	}

	return
}

func getIP6(domain string) (res []string) {
	if r, _, err := GetRR(domain, []string{"127.0.0.11:53"}, dns.TypeAAAA); err == nil {
		switch r.Rcode {
		case dns.RcodeSuccess:
			if len(r.Answer) > 0 {
				if len(r.Answer) > 99 {
					Warning.Printf("Internal error, for %s, Answer too big: %d\n", domain, len(r.Answer))
				}
				cnt := 0
				for _, rr := range r.Answer {
					switch rr.Header().Rrtype {
					case dns.TypeAAAA:
						if cnt > 8 {
							continue
						}
						res = appendIfMissing(res, rr.(*dns.AAAA).AAAA.String())
						cnt++
					case dns.TypeCNAME:
						res1 := getIP6(strings.TrimSuffix(rr.(*dns.CNAME).Target, "."))
						for _, ele := range res1 {
							res = appendIfMissing(res, ele)
						}
					case dns.TypeRRSIG:
						Debug.Printf("Warning: RRSIG (%s): %#v", domain, r.MsgHdr)
					default:
						Debug.Printf("Warning: unknown answer (%s): %s\n", domain, rr.String())
					}
				}
			}
		}
	} else {
		Error.Printf("Type AAAA. Internal error (%s): %s\n", domain, err.Error())
	}
	return
}

func GetRR(domain string, nameservers []string, qtype uint16) (r *dns.Msg, rtt time.Duration, err error) {
	if len(nameservers) == 0 {
		err = fmt.Errorf("%s", "No nameservers!")
		return
	}
	for a := 0; a < ATTEMPTS; a++ {
		if a > 1 {
			time.Sleep(250 * time.Millisecond)
		}
		l := rand.Perm(len(nameservers))
		for _, i := range l {
			nameserver := nameservers[i]
			m := &dns.Msg{
				MsgHdr: dns.MsgHdr{
					//Authoritative: true,
					AuthenticatedData: true,
					CheckingDisabled:  true,
					RecursionDesired:  true,
					Opcode:            dns.OpcodeQuery,
					Rcode:             dns.RcodeSuccess,
				},
				Question: make([]dns.Question, 1),
			}
			o := &dns.OPT{
				Hdr: dns.RR_Header{
					Name:   ".",
					Rrtype: dns.TypeOPT,
				},
			}
			o.SetDo()
			o.SetUDPSize(dns.DefaultMsgSize)
			m.Extra = append(m.Extra, o)
			qt := qtype
			qc := uint16(dns.ClassINET)
			m.Question[0] = dns.Question{Name: dns.Fqdn(domain), Qtype: qt, Qclass: qc}
			m.Id = dns.Id()
			r, rtt, err = lookup(m, nameserver, true)
			if err == nil {
				break
			}
		}
		if err == nil {
			break
		}
	}
	return r, rtt, err
}

func lookup(m *dns.Msg, nameserver string, fallback bool) (r *dns.Msg, rtt time.Duration, err error) {
	c := new(dns.Client)
	c.Timeout = time.Second * TIMEOUT
	if fallback {
		c.Net = "udp"
	} else {
		c.Net = "tcp"
	}
	r, rtt, err = c.Exchange(m, nameserver)
	/*
		switch err {
		case nil:
			//do nothing
			                case dns.ErrTruncated:
						if fallback {
							// First EDNS, then TCP
							c.Net = "tcp"
							r, rtt, err = lookup(m, nameserver, false)
						}
		default:
			//do nothing
		}
	*/
	if r != nil {
		if r.Truncated {
			if fallback {
				// First EDNS, then TCP
				c.Net = "tcp"
				r, rtt, err = lookup(m, nameserver, false)
			}
		}
	}
	if err == nil {
		if r.Id != m.Id {
			err = fmt.Errorf("%s", "Id mismatch")
		}
		//fmt.Printf("%v", r)
		//fmt.Printf("\n;; query time: %.3d µs, server: %s(%s), size: %d bytes\n", rtt/1e3, nameserver, c.Net, r.Len())
	}
	if err != nil && r != nil {
		if m.Response || m.Opcode == dns.OpcodeQuery {
			err = nil
		}
	}
	return r, rtt, err
}
