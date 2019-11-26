package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	pb "github.com/usher2/u2ckbot/msg"
)

const (
	TBLOCK_URL = iota
	TBLOCK_HTTPS
	TBLOCK_DOMAIN
	TBLOCK_MASK
	TBLOCK_IP
)

const PRINT_LIMIT = 10

type TReason struct {
	Id     int32
	Aggr   []string
	Ip     []string
	Url    []string
	Domain []string
}

func constructContentResult(a []*pb.Content) (res string) {
	var mass string
	var oldest int64 = 1<<63 - 1
	if len(a) == 0 {
		return
	}
	for _, packet := range a {
		content := TContent{}
		err := json.Unmarshal(packet.Pack, &content)
		if err != nil {
			Error.Printf("Oooops!!! %s\n", err.Error)
			continue
		}
		if packet.RegistryUpdateTime < oldest {
			oldest = packet.RegistryUpdateTime
		}
		Debug.Printf("%v\n", content)
		if len(content.Subnet4)+len(content.Subnet6) > 0 && packet.BlockType == TBLOCK_IP {
			mass = "\U0001f4a5\U0001f4a5\U0001f4a5 Decision for mass blocking!\n\n"
		}
		bt := ""
		switch packet.BlockType {
		case TBLOCK_URL:
			bt = "\U000026d4 (url) "
		case TBLOCK_HTTPS:
			bt = "\U0001f4db (https) "
		case TBLOCK_DOMAIN:
			bt = "\U0001f6ab (domain) "
		case TBLOCK_MASK:
			bt = "\U0001f506 (wildcard) "
		case TBLOCK_IP:
			bt = "\u274c (ip) "
		}
		dcs := fmt.Sprintf("%s %s %s", content.Decision.Org, content.Decision.Number, content.Decision.Date)
		res += fmt.Sprintf("%s #%d %s\n", bt, content.Id, dcs)
		res += fmt.Sprintf("included: %s\n", time.Unix(content.IncludeTime, 0).In(time.FixedZone("UTC+3", 3*60*60)).Format(time.RFC3339))
		for i, d := range content.Domain {
			if i >= PRINT_LIMIT {
				res += fmt.Sprintf("... and %d other domains\n", len(content.Domain)-i)
				break
			}
			res += fmt.Sprintf("  domain: %s\n", Sanitize(d.Domain))
		}
		for i, u := range content.Url {
			if i >= PRINT_LIMIT {
				res += fmt.Sprintf("... and %d other urls\n", len(content.Url)-i)
				break
			}
			res += fmt.Sprintf("  url: %s\n", Sanitize(u.Url))
		}
		for i, ip := range content.Ip4 {
			if i >= PRINT_LIMIT {
				res += fmt.Sprintf("... and %d other ips\n", len(content.Ip4)-i)
				break
			}
			ip4 := fmt.Sprintf("%d.%d.%d.%d",
				(ip.Ip4&0xFF000000)>>24,
				(ip.Ip4&0x00FF0000)>>16,
				(ip.Ip4&0x0000FF00)>>8,
				(ip.Ip4 & 0x000000FF),
			)
			res += fmt.Sprintf("  IP: %s\n", ip4)
		}
		for i, ip := range content.Ip6 {
			if i >= PRINT_LIMIT {
				res += fmt.Sprintf("... and %d other ips\n", len(content.Ip6)-i)
				break
			}
			res += fmt.Sprintf("  IP: %s\n", net.IP(ip.Ip6).String())
		}
		for i, sb := range content.Subnet4 {
			if i >= PRINT_LIMIT {
				res += fmt.Sprintf("... and %d other subnets\n", len(content.Subnet4)-i)
				break
			}
			res += fmt.Sprintf("  Subnet: %s\n", sb.Subnet4)
		}
		for i, sb := range content.Subnet6 {
			if i >= PRINT_LIMIT {
				res += fmt.Sprintf("... and %d other subnets\n", len(content.Subnet6)-i)
				break
			}
			res += fmt.Sprintf("  Subnet: %s\n", sb.Subnet6)
		}
		break
	}
	if mass == "" {
		res = "\n" + res
	} else {
		res = mass + res
	}
	res += fmt.Sprintf("\n_relevance not earlier than:_ %s\n", time.Unix(oldest, 0).In(time.FixedZone("UTC+3", 3*60*60)).Format(time.RFC3339))
	return
}

func constructResult(a []*pb.Content) (res string) {
	var mass string
	var oldest int64 = 1<<63 - 1
	var ra []TReason
	if len(a) == 0 {
		return
	}
	sort.Slice(a, func(i, j int) bool {
		return a[i].Id < a[j].Id
	})
	ra = make([]TReason, 1)
	ra[0].Id = a[0].Id
	if a[0].Aggr != "" {
		ra[0].Aggr = append(ra[0].Aggr, strings.Split(a[0].Aggr, ",")...)
	}
	if a[0].Ip4 != 0 {
		ra[0].Ip = append(ra[0].Ip, fmt.Sprintf("%d.%d.%d.%d",
			(a[0].Ip4&0xFF000000)>>24,
			(a[0].Ip4&0x00FF0000)>>16,
			(a[0].Ip4&0x0000FF00)>>8,
			(a[0].Ip4&0x000000FF),
		))
	}
	if len(a[0].Ip6) != 0 {
		ra[0].Ip = append(ra[0].Ip, net.IP(a[0].Ip6).String())
	}
	if a[0].Domain != "" {
		ra[0].Domain = append(ra[0].Domain, PrintedDomain(a[0].Domain))
	}
	if a[0].Url != "" {
		ra[0].Url = append(ra[0].Url, PrintedDomain(a[0].Url))
	}
	for i := 0; i < len(a)-1; i++ {
		if a[i].Id == a[i+1].Id {
			if a[i+1].Aggr != "" {
				ra[i].Aggr = append(ra[i].Aggr, strings.Split(a[i+1].Aggr, ",")...)
			}
			if a[i+1].Ip4 != 0 {
				ra[i].Ip = append(ra[i].Ip, fmt.Sprintf("%d.%d.%d.%d",
					(a[i+1].Ip4&0xFF000000)>>24,
					(a[i+1].Ip4&0x00FF0000)>>16,
					(a[i+1].Ip4&0x0000FF00)>>8,
					(a[i+1].Ip4&0x000000FF),
				))
			}
			if len(a[i+1].Ip6) != 0 {
				ra[i].Ip = append(ra[i].Ip, net.IP(a[i+1].Ip6).String())
			}
			if a[i+1].Domain != "" {
				ra[i].Domain = append(ra[i].Domain, a[i+1].Domain)
			}
			if a[i+1].Url != "" {
				ra[i].Url = append(ra[i].Url, a[i+1].Url)
			}
			a = append(a[:i], a[i+1:]...)
			i--
		} else {
			ra = append(ra, TReason{})
			ra[i+1].Id = a[i+1].Id
			if a[i+1].Aggr != "" {
				ra[i+1].Aggr = append(ra[i+1].Aggr, strings.Split(a[i+1].Aggr, ",")...)
			}
			if a[i+1].Ip4 != 0 {
				ra[i+1].Ip = append(ra[i+1].Ip, fmt.Sprintf("%d.%d.%d.%d",
					(a[i+1].Ip4&0xFF000000)>>24,
					(a[i+1].Ip4&0x00FF0000)>>16,
					(a[i+1].Ip4&0x0000FF00)>>8,
					(a[i+1].Ip4&0x000000FF),
				))
			}
			if len(a[i+1].Ip6) != 0 {
				ip6 := net.IP(a[i+1].Ip6)
				ra[i+1].Ip = append(ra[i+1].Ip, ip6.String())
			}
			if a[i+1].Domain != "" {
				ra[i+1].Domain = append(ra[i+1].Domain, a[i+1].Domain)
			}
			if a[i+1].Url != "" {
				ra[i+1].Url = append(ra[i+1].Url, a[i+1].Url)
			}

		}
	}
	sort.Slice(a, func(j, i int) bool {
		if a[i].BlockType == TBLOCK_URL && a[j].BlockType != TBLOCK_URL {
			return true
		} else if a[i].BlockType == TBLOCK_HTTPS &&
			(a[j].BlockType != TBLOCK_URL &&
				a[j].BlockType != TBLOCK_HTTPS) {
			return true
		} else if a[i].BlockType == TBLOCK_DOMAIN &&
			(a[j].BlockType != TBLOCK_URL &&
				a[j].BlockType != TBLOCK_HTTPS &&
				a[j].BlockType != TBLOCK_DOMAIN) {
			return true
		} else if a[i].BlockType == TBLOCK_MASK &&
			(a[j].BlockType != TBLOCK_URL &&
				a[j].BlockType != TBLOCK_HTTPS &&
				a[j].BlockType != TBLOCK_DOMAIN &&
				a[j].BlockType != TBLOCK_MASK) {
			return true
		} else {
			return false
		}
	})
	var cnt, cbu, cbh, cbd, cbm, cbi int
	for _, packet := range a {
		content := TContent{}
		err := json.Unmarshal(packet.Pack, &content)
		if err != nil {
			Error.Printf("Oooops!!! %s\n", err.Error)
			continue
		}
		if packet.RegistryUpdateTime < oldest {
			oldest = packet.RegistryUpdateTime
		}
		var req TReason
		for _, req = range ra {
			if req.Id == packet.Id {
				break
			}
		}
		if len(req.Aggr) != 0 {
			if packet.BlockType == TBLOCK_IP {
				mass = "\U0001f4a5\U0001f4a5\U0001f4a5 Mass blocked resource!\n\n"
			}
		}
		if cnt < PRINT_LIMIT {
			bt := ""
			switch packet.BlockType {
			case TBLOCK_URL:
				bt = "\U000026d4 "
				cbu++
			case TBLOCK_HTTPS:
				bt = "\U0001f4db "
				cbh++
			case TBLOCK_DOMAIN:
				bt = "\U0001f6ab "
				cbd++
			case TBLOCK_MASK:
				bt = "\U0001f506 "
				cbm++
			case TBLOCK_IP:
				bt = "\u274c "
				cbi++
			}
			dcs := fmt.Sprintf("%s %s %s", content.Decision.Org, content.Decision.Number, content.Decision.Date)
			res += fmt.Sprintf("%s #%d %s\n", bt, content.Id, dcs)
			if len(req.Aggr) != 0 {
				for _, nw := range req.Aggr {
					res += fmt.Sprintf("    _as subnet_ %s\n", nw)
				}
			}
			if len(req.Ip) != 0 {
				for _, ip := range req.Ip {
					res += fmt.Sprintf("    _as ip_ %s\n", ip)
				}
			}
			if len(req.Domain) != 0 {
				for _, domain := range req.Domain {
					res += fmt.Sprintf("    _as domain_ %s\n", Sanitize(PrintedDomain(domain)))
				}
			}
			if len(req.Url) != 0 {
				for _, u := range ra[packet.Id].Url {
					res += fmt.Sprintf("    _as url_ %s\n", Sanitize(PrintedDomain(u)))
				}
			}
			res += "\n"
		}
		cnt++
	}
	if mass != "" {
		res = mass + res
	}
	if cnt > PRINT_LIMIT {
		rest := cnt - PRINT_LIMIT
		res += fmt.Sprintf("\U000026f0 ... and %d others\n", rest)
		/*if cbu > 0 && cbu < rest {
			res += fmt.Sprintf(" url=%d", cbu)
		} else if cbu > 0 {
			res += fmt.Sprintf(" url=%d", rest)
		}
		if cbh > 0 && cbu+cbh < rest {
			res += fmt.Sprintf(" https=%d", cbu)
		} else if cbh > 0 {
			res += fmt.Sprintf(" https=%d", rest-cbu)
		}
		if cbd > 0 && cbd+cbu+cbh < rest {
			res += fmt.Sprintf(" domain=%d", cbu)
		} else if cbd > 0 {
			res += fmt.Sprintf(" domain=%d", rest-cbh-cbu)
		}
		if cbm > 0 && cbm+cbd+cbu+cbh < rest {
			res += fmt.Sprintf(" wildcard=%d", cbu)
		} else if cbm > 0 {
			res += fmt.Sprintf(" wildcard=%d", rest-cbd-cbh-cbu)
		}
		if cbi > 0 && cbm+cbd+cbu+cbh < rest {
			res += fmt.Sprintf(" ip=%d", rest-cbm-cbd-cbh-cbu)
		}*/
		res += "\n"
	}
	var abt []string
	if cbu > 0 {
		abt = append(abt, "url: \U000026d4")
	}
	if cbh > 0 {
		abt = append(abt, "https: \U0001f4db")
	}
	if cbd > 0 {
		abt = append(abt, "domain: \U0001f6ab")
	}
	if cbm > 0 {
		abt = append(abt, "wildcard: \U0001f506")
	}
	if cbi > 0 {
		abt = append(abt, "ip: \u274c")
	}
	res += "*blocking types:* " + strings.Join(abt, " | ")
	res += fmt.Sprintf("\n_relevance not earlier than:_ %s\n", time.Unix(oldest, 0).In(time.FixedZone("UTC+3", 3*60*60)).Format(time.RFC3339))
	return
}

func searchID(c pb.CheckClient, id string) ([]*pb.Content, error) {
	fmt.Printf("Looking for content: %s\n", id)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	id32, _ := strconv.Atoi(id)
	r, err := c.SearchID(ctx, &pb.IDRequest{Query: int32(id32)})
	if err != nil {
		Debug.Printf("%v.SearchContent(_) = _, %v\n", c, err)
		return nil, fmt.Errorf("\U00002620 Something wrong! Try again later\n")
	}
	if r.Error != "" {
		Debug.Printf("ERROR: %s\n", r.Error)
		return nil, fmt.Errorf("\u23f3 Try again later: %s\n", r.Error)
	}
	return r.Results[:], nil
}

func searchIP4(c pb.CheckClient, ip string) ([]*pb.Content, error) {
	fmt.Printf("Looking for %s\n", ip)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := c.SearchIP4(ctx, &pb.IP4Request{Query: parseIp4(ip)})
	if err != nil {
		Debug.Printf("%v.SearchIP4(_) = _, %v\n", c, err)
		return nil, fmt.Errorf("\U00002620 Something wrong! Try again later\n")
	}
	if r.Error != "" {
		Debug.Printf("ERROR: %s\n", r.Error)
		return nil, fmt.Errorf("\u23f3 Try again later: %s\n", r.Error)
	}
	return r.Results[:], nil
}

func searchIP6(c pb.CheckClient, ip string) ([]*pb.Content, error) {
	fmt.Printf("Looking for %s\n", ip)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ip6 := net.ParseIP(ip)
	if len(ip6) == 0 {
		return nil, fmt.Errorf("Can't parse IP: %s\n", ip)
	}
	r, err := c.SearchIP6(ctx, &pb.IP6Request{Query: ip6})
	if err != nil {
		Debug.Printf("%v.SearchIP6(_) = _, %v\n", c, err)
		return nil, fmt.Errorf("\U00002620 Something wrong! Try again later\n")
	}
	if r.Error != "" {
		Debug.Printf("ERROR: %s\n", r.Error)
		return nil, fmt.Errorf("\u23f3 Try again later: %s\n", r.Error)
	}
	return r.Results[:], nil
}

func searchURL(c pb.CheckClient, u string) ([]*pb.Content, error) {
	_url := NormalizeUrl(u)
	if _url != u {
		fmt.Printf("Input was %s\n", u)
	}
	fmt.Printf("Looking for %s\n", _url)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := c.SearchURL(ctx, &pb.URLRequest{Query: _url})
	if err != nil {
		Debug.Printf("%v.SearchURL(_) = _, %v\n", c, err)
		return nil, fmt.Errorf("\U00002620 Something wrong! Try again later\n")
	}
	if r.Error != "" {
		Debug.Printf("ERROR: %s\n", r.Error)
		return nil, fmt.Errorf("\u23f3 Try again later: %s\n", r.Error)
	}
	return r.Results[:], nil
}

func searchDomain(c pb.CheckClient, s string) ([]*pb.Content, error) {
	domain := NormalizeDomain(s)
	Debug.Printf("Looking for %s\n", domain)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	r, err := c.SearchDomain(ctx, &pb.DomainRequest{Query: domain})
	if err != nil {
		Debug.Printf("%v.SearchURL(_) = _, %v\n", c, err)
		return nil, fmt.Errorf("\U00002620 Something wrong! Try again later\n")
	}
	if r.Error != "" {
		Debug.Printf("ERROR: %s\n", r.Error)
		return nil, fmt.Errorf("\u23f3 Try again later: %s\n", r.Error)
	}
	return r.Results[:], nil
}

func refSearch(c pb.CheckClient, s string) ([]*pb.Content, []string, []string, error) {
	var err error
	var a, a2 []*pb.Content
	var ips4, ips6 []string
	domain := NormalizeDomain(s)
	ips4 = getIP4(domain)
	for _, ip := range ips4 {
		a2, err = searchIP4(c, ip)
		if err == nil {
			a = append(a, a2...)
		} else {
			break
		}
	}
	if err == nil {
		ips6 = getIP6(domain)
		for _, ip := range ips6 {
			a2, err = searchIP6(c, ip)
			if err == nil {
				a = append(a, a2...)
			} else {
				break
			}
		}
	}
	if err != nil {
		return nil, ips4, ips6, err
	} else {
		return a, ips4, ips6, nil
	}
}

func mainSearch(c pb.CheckClient, s string) (res string) {
	var err error
	var a, a2 []*pb.Content
	domain := NormalizeDomain(s)
	if len(s) > 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-2]
			domain = s
		}
	}
	_u, _ur := url.Parse(s)
	if _ur == nil && _u.IsAbs() &&
		(_u.Scheme == "http" || _u.Scheme == "https") &&
		(_u.Port() == "80" || _u.Port() == "443" || _u.Port() == "") &&
		(_u.RequestURI() == "" || _u.RequestURI() == "/") {
		s = _u.Hostname()
		domain = NormalizeDomain(s)
		_ur = fmt.Errorf("fake")
	}
	ip := net.ParseIP(s)
	_c, _ := strconv.Atoi(s)
	if ip != nil {
		if ip.To4() != nil {
			a, err = searchIP4(c, s)
			if err == nil {
				a2, err = searchDomain(c, s)
				if err == nil {
					if len(a2) > 0 {
						a = append(a, a2...)
					}
				}
			}
		} else {
			a, err = searchIP6(c, s)
		}
		if err == nil {
			if len(a) > 0 {
				res = fmt.Sprintf("\U0001f525 %s *is blocked*\n\n", Sanitize(s))
			} else {
				res = fmt.Sprintf("\u2705 %s *is not blocked*\n", Sanitize(s))
			}
		}
		if err != nil {
			res = err.Error() + "\n"
		} else {
			res += constructResult(a)
		}
	} else if isDomainName(domain) {
		a, err = searchDomain(c, s)
		if err == nil {
			if strings.HasPrefix(s, "www.") {
				a2, err = searchDomain(c, s[4:])
			} else {
				a2, err = searchDomain(c, "www."+s)
			}

		}
		if err == nil {
			if len(a2) > 0 {
				a = append(a, a2...)
			}
			if len(a) > 0 {
				res = fmt.Sprintf("\U0001f525 %s *is blocked*\n\n", Sanitize(s))
			} else {
				res = fmt.Sprintf("\u2705 %s *is not blocked*\n", Sanitize(s))
				var ips4, ips6 []string
				a, ips4, ips6, err = refSearch(c, s)
				if err == nil && len(a) > 0 {
					res += fmt.Sprintf("\n\U0001f525 but may be filtered by IP:\n")
					for _, ip := range ips4 {
						res += fmt.Sprintf("    %s\n", ip)
					}
					for _, ip := range ips6 {
						res += fmt.Sprintf("    %s\n", ip)
					}
					res += "\n"
				}
			}
		}
		if err != nil {
			res = err.Error() + "\n"
		} else {
			res += constructResult(a)
		}
	} else if _c != 0 {
		a, err = searchID(c, s)
		if err == nil {
			if len(a) == 0 {
				res = fmt.Sprintf("\U0001f914 %s *is not found*\n", s)
			}
		}
		if err != nil {
			res = err.Error() + "\n"
		} else {
			res += constructContentResult(a)
		}
	} else if s[0] == '#' {
		_, err = strconv.Atoi(string(s[1:]))
		if err == nil {
			a, err = searchID(c, s[1:])
			if err == nil {
				if len(a) == 0 {
					res = fmt.Sprintf("\U0001f914 %s *is not found*\n", s)
				}
			}
		}
		if err != nil {
			res = fmt.Sprintf("\U0001f914 What did you mean?.. %s\n", s)
		} else {
			res += constructContentResult(a)
		}
	} else if _ur == nil {
		if _u.Scheme != "https" && _u.Scheme != "http" {
			a, err = searchURL(c, s)
		} else {
			_u.Scheme = "https"
			a, err = searchURL(c, _u.String())
			if err == nil {
				_u.Scheme = "http"
				a2, err = searchURL(c, _u.String())
				if err == nil {
					if len(a2) > 0 {
						a = append(a, a2...)
					}
				}
			}
		}
		if err == nil {
			if len(a) > 0 {
				res = fmt.Sprintf("\U0001f525 %s *is blocked*\n\n", Sanitize(s))
			} else {
				res = fmt.Sprintf("\u2705 %s *is not blocked*\n", Sanitize(s))
			}
		}
		if err != nil {
			res = err.Error() + "\n"
		} else {
			res += constructResult(a)
		}
	} else {
		a, err = searchURL(c, s)
		if err == nil {
			a2, err = searchDomain(c, s)
			if err == nil {
				if len(a2) > 0 {
					a = append(a, a2...)
				}
			}
			if len(a) > 0 {
				res += constructResult(a)
			} else {
				res = fmt.Sprintf("\U0001f914 What did you mean?.. %s\n", s)
			}
		}
		if err != nil {
			res = err.Error() + "\n"
		}
	}
	return
}
