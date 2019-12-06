package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	pb "github.com/usher2/u2ckbot/msg"
)

func searchID(c pb.CheckClient, id string) ([]*pb.Content, error) {
	Info.Printf("Looking for content: %s\n", id)
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
	Info.Printf("Looking for %s\n", ip)
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
	Info.Printf("Looking for %s\n", ip)
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
	Info.Printf("Looking for %s\n", _url)
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
	Info.Printf("Looking for %s\n", domain)
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
	if len(s) == 0 {
		res = fmt.Sprintf("\U0001f914 What did you mean?..\n")
		return
	}
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
		_, err = strconv.Atoi(s[1:])
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
				res = fmt.Sprintf("\U0001f525 URL %s *is blocked*\n\n", Sanitize(s))
			} else {
				res = fmt.Sprintf("\u2705 URL %s *is not blocked*\n", Sanitize(s))
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
		}
		if err != nil {
			res = err.Error() + "\n"
		} else {
			if len(a) > 0 {
				res = fmt.Sprintf("\U0001f525 URL %s *is blocked*\n\n", Sanitize(s))
				res += constructResult(a)
			} else {
				res = fmt.Sprintf("\U0001f914 What did you mean?.. %s\n", s)
			}
		}
	}
	return
}
