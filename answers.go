package main

import (
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/usher2/u2ckbot/internal/logger"
	pb "github.com/usher2/u2ckbot/msg"
)

const (
	TBLOCK_URL = iota
	TBLOCK_HTTPS
	TBLOCK_DOMAIN
	TBLOCK_MASK
	TBLOCK_IP
)

const encodeCorc = "abcdefghjkmnpqrstvwxyz0123456789"

const PRINT_LIMIT = 5

const MAX_TIMESTAMP int64 = 1<<63 - 1

const (
	DTIME12 = 12 * 60 * 60
	DTIME3  = 3 * 60 * 60
)

const (
	_ = iota
	OFFSET_CONTENT
	OFFSET_IP4
	OFFSET_IP6
	OFFSET_URL
	OFFSET_SUBNET4
	OFFSET_SUBNET6
	OFFSET_DOMAIN
)

type TPagination struct {
	Tag, Count int
}

type TReason struct {
	Id     int32
	Aggr   []string
	Ip     []string
	Url    []string
	Domain []string
}

var DecisionTypesDesc = map[string]string{
	"15_1_1m": "[ст. 15.1 п.1(м)](http://www.consultant.ru/document/cons_doc_LAW_61798/38c8ea666d27d9dc12b078c556e316e90248f551/), каталог VPN",
	"15_1":    "[ст. 15.1](http://www.consultant.ru/document/cons_doc_LAW_61798/38c8ea666d27d9dc12b078c556e316e90248f551/), общая",
	"15_1_1":  "[ст. 15.1-1](http://www.consultant.ru/document/cons_doc_LAW_61798/079aac275ffc6cea954b19c5b177a547b94f3c48/), неуважение",
	"15_2":    "[ст. 15.2](http://www.consultant.ru/document/cons_doc_LAW_61798/1f316dc4a18023edcd030bc6591c4dd8b4f841dc/), правообладание",
	"15_3":    "[ст. 15.3](http://www.consultant.ru/document/cons_doc_LAW_61798/34547c9b6ddb60cebd0a67593943fd9ef64ebdd0/), мятеж и фейки",
	"15_4":    "[ст. 15.4](http://www.consultant.ru/document/cons_doc_LAW_61798/96723dcd9be73473a978013263f16f42cd8cd53d/), ОРИ не молчи",
	"15_6":    "[ст. 15.6](http://www.consultant.ru/document/cons_doc_LAW_61798/c7c4ad36689c46c7e8a3ab49c9db8ccbc7c82920/), вечная",
	"15_6_1":  "[ст. 15.6-1](http://www.consultant.ru/document/cons_doc_LAW_61798/c7c4ad36689c46c7e8a3ab49c9db8ccbc7c82920/), вечная зеркал",
	"15_5":    "[ст. 15.5](http://www.consultant.ru/document/cons_doc_LAW_61798/98228cbe6565abbe55d0842a7e8593012c3449ea/), персональные данные",
	"15_8":    "[ст. 15.8](http://www.consultant.ru/document/cons_doc_LAW_61798/1a807328c80a540bd0bb724927d6e774595431dc/), VPN",
	"15_9":    "[ст. 15.9](http://www.consultant.ru/document/cons_doc_LAW_61798/31eb19e991d54b484ac546107c4db838b3631e9f/), сайт иноагента",
}

func DecisionTypeView(t string) string {
	t = strings.ReplaceAll(strings.ReplaceAll(t, ".", "_"), "-", "_")

	if val, ok := DecisionTypesDesc[t]; ok {
		return val
	}

	return DecisionTypesDesc["15_1"]
}

var BlockTypesView = map[int32]string{
	TBLOCK_URL:    "\U000026d4 (url) ",
	TBLOCK_HTTPS:  "\U0001f4db (https) ",
	TBLOCK_DOMAIN: "\U0001f6ab (domain) ",
	TBLOCK_MASK:   "\U0001f506 (wildcard) ",
	TBLOCK_IP:     "\u274c (ip) ",
}

func BlockTypeView(t int32) string {
	if val, ok := BlockTypesView[t]; ok {
		return val
	}

	return BlockTypesView[TBLOCK_URL]
}

func printUpToDate(t int64) string {
	var r rune
	d := time.Now().Unix() - t
	switch {
	case d > DTIME12:
		r = 0x2b55
	case d > DTIME3:
		r = 0x000026a0
	default:
		r = 0x2705
	}
	return fmt.Sprintf("\n%c _Данные синхронизированы:_ %s\n", r, time.Unix(t, 0).In(time.FixedZone("UTC+3", 3*60*60)).Format(time.RFC3339))
}

func Uint64ToBase32(i uint64) string {
	b32 := base32.NewEncoding(encodeCorc).WithPadding(base32.NoPadding)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return b32.EncodeToString(b)
}

func Base32ToUint64(s string) (uint64, error) {
	b32 := base32.NewEncoding(encodeCorc).WithPadding(base32.NoPadding)
	b, err := b32.DecodeString(s)
	if err == nil {
		return binary.LittleEndian.Uint64(b), nil
	}
	return 0, err
}

func String2fnv2base32(s string) string {
	h64 := fnv.New64a()
	h64.Write([]byte(s))
	return Uint64ToBase32(h64.Sum64())
}

func constructBasis(entryType int32, org, number string) string {
	switch {
	case entryType == 1 && org == "Роскомнадзор" && strings.HasSuffix(number, "-СОБ"):
		return DecisionTypeView("15.1.1m")
	case entryType == 1 && (org == "Генпрокуратура" || org == ""):
		return DecisionTypeView("15.1-1")
	case entryType == 2:
		return DecisionTypeView("15.2")
	case entryType == 3:
		return DecisionTypeView("15.3")
	case entryType == 4:
		return DecisionTypeView("15.4")
	case entryType == 5 && org == "Мосгорсуд":
		return DecisionTypeView("15.6")
	case entryType == 5 && (org == "Минцифра" || org == "Минкомсвязь"):
		return DecisionTypeView("15.6-1")
	case entryType == 6:
		return DecisionTypeView("15.5")
	case entryType == 7:
		return DecisionTypeView("15.8")
	case entryType == 8:
		return DecisionTypeView("15.9")
	}

	return DecisionTypeView("15.1")
}

func constructContentResult(a []*pb.Content, o TPagination) (res string, pages []TPagination) {
	var oldest int64 = MAX_TIMESTAMP

	if len(a) == 0 {
		return
	}

	for _, packet := range a {
		content := Content{}

		if err := json.Unmarshal(packet.Pack, &content); err != nil {
			logger.Error.Printf("Упс!!! %s\n", err)

			continue
		}

		if packet.RegistryUpdateTime < oldest {
			oldest = packet.RegistryUpdateTime
		}

		blockType := BlockTypeView(packet.BlockType)

		descisionString := fmt.Sprintf("%s %s %s", content.Decision.Org, content.Decision.Number, content.Decision.Date)
		res += fmt.Sprintf("%s /n\\_%d %s /d\\_%s\n", blockType, content.ID, descisionString, String2fnv2base32(descisionString))
		res += fmt.Sprintf("\u2022 %s\n", constructBasis(content.EntryType, printOrg(content.Decision.Org), content.Decision.Number))
		res += fmt.Sprintf("внесено: %s\n", time.Unix(content.IncludeTime, 0).In(time.FixedZone("UTC+3", 3*60*60)).Format(time.RFC3339))

		if len(content.SubnetIPv4)+len(content.SubnetIPv6) > 0 && packet.BlockType == TBLOCK_IP {
			res += "\U0001f4a5\U0001f4a5\U0001f4a5 Решение о «ковровой» блокировке!\n"
		}

		res += "\n"

		cnt := 0

		for i, d := range content.Domain {
			if o.Tag == OFFSET_DOMAIN && i < o.Count {
				continue
			}

			if cnt >= PRINT_LIMIT {
				break
			}

			res += fmt.Sprintf("  domain: %s\n", Sanitize(d.Domain))

			cnt++
		}

		l := len(content.Domain)
		if l > PRINT_LIMIT {
			offset := 0
			if o.Tag == OFFSET_DOMAIN {
				switch {
				case l <= PRINT_LIMIT:
					offset = 0
				case o.Count > l-(l%PRINT_LIMIT):
					offset = l - (l % PRINT_LIMIT)
				default:
					offset = o.Count
				}
			}

			res += fmt.Sprintf("  \u2195 результаты с *%d* по *%d* из *%d*\n", offset+1, offset+cnt, l)

			pages = append(pages, TPagination{OFFSET_DOMAIN, l})
		}

		if cnt > 0 {
			res += "\n"
		}

		cnt = 0
		for i, u := range content.URL {
			if o.Tag == OFFSET_URL && i < o.Count {
				continue
			}
			if cnt >= PRINT_LIMIT {
				break
			}
			res += fmt.Sprintf("  url: %s\n", Sanitize(u.URL))
			cnt++
		}
		l = len(content.URL)
		if l > PRINT_LIMIT {
			offset := 0
			if o.Tag == OFFSET_URL {
				switch {
				case l <= PRINT_LIMIT:
					offset = 0
				case o.Count > l-(l%PRINT_LIMIT):
					offset = l - (l % PRINT_LIMIT)
				default:
					offset = o.Count
				}
			}
			res += fmt.Sprintf("  \u2195 результаты с *%d* по *%d* из *%d*\n", offset+1, offset+cnt, l)
			pages = append(pages, TPagination{OFFSET_URL, l})
		}
		if cnt > 0 {
			res += "\n"
		}
		cnt = 0
		for i, ip := range content.IPv4 {
			if o.Tag == OFFSET_IP4 && i < o.Count {
				continue
			}
			if cnt >= PRINT_LIMIT {
				break
			}
			res += fmt.Sprintf("  IP: %s\n", int2Ip4(ip.IPv4))
			cnt++
		}
		l = len(content.IPv4)
		if l > PRINT_LIMIT {
			offset := 0
			if o.Tag == OFFSET_IP4 {
				switch {
				case l <= PRINT_LIMIT:
					offset = 0
				case o.Count > l-(l%PRINT_LIMIT):
					offset = l - (l % PRINT_LIMIT)
				default:
					offset = o.Count
				}
			}
			res += fmt.Sprintf("  \u2195 результаты с *%d* по *%d* из *%d*\n", offset+1, offset+cnt, l)
			pages = append(pages, TPagination{OFFSET_IP4, l})
		}
		if cnt > 0 {
			res += "\n"
		}
		cnt = 0
		for i, ip := range content.IPv6 {
			if o.Tag == OFFSET_IP6 && i < o.Count {
				continue
			}
			if cnt >= PRINT_LIMIT {
				break
			}
			res += fmt.Sprintf("  IP: %s\n", net.IP(ip.IPv6).String())
			cnt++
		}
		l = len(content.IPv6)
		if l > PRINT_LIMIT {
			offset := 0
			if o.Tag == OFFSET_IP6 {
				switch {
				case l <= PRINT_LIMIT:
					offset = 0
				case o.Count > l-(l%PRINT_LIMIT):
					offset = l - (l % PRINT_LIMIT)
				default:
					offset = o.Count
				}
			}
			res += fmt.Sprintf("  \u2195 результаты с *%d* по *%d* из *%d*\n", offset+1, offset+cnt, l)
			pages = append(pages, TPagination{OFFSET_IP6, l})
		}
		if cnt > 0 {
			res += "\n"
		}
		cnt = 0
		for i, sb := range content.SubnetIPv4 {
			if o.Tag == OFFSET_SUBNET4 && i < o.Count {
				continue
			}
			if cnt >= PRINT_LIMIT {
				break
			}
			res += fmt.Sprintf("  Подсеть: %s\n", sb.SubnetIPv4)
			cnt++
		}
		l = len(content.SubnetIPv4)
		if l > PRINT_LIMIT {
			offset := 0
			if o.Tag == OFFSET_SUBNET4 {
				switch {
				case l <= PRINT_LIMIT:
					offset = 0
				case o.Count > l-(l%PRINT_LIMIT):
					offset = l - (l % PRINT_LIMIT)
				default:
					offset = o.Count
				}
			}
			res += fmt.Sprintf("  \u2195 результаты с *%d* по *%d* из *%d*\n", offset+1, offset+cnt, l)
			pages = append(pages, TPagination{OFFSET_SUBNET4, l})
		}
		if cnt > 0 {
			res += "\n"
		}
		cnt = 0
		for i, sb := range content.SubnetIPv6 {
			if o.Tag == OFFSET_SUBNET6 && i < o.Count {
				continue
			}
			if cnt >= PRINT_LIMIT {
				break
			}
			res += fmt.Sprintf("  Подсеть: %s\n", sb.SubnetIPv6)
			cnt++
		}
		l = len(content.SubnetIPv6)
		if l > PRINT_LIMIT {
			offset := 0
			if o.Tag == OFFSET_SUBNET6 {
				switch {
				case l <= PRINT_LIMIT:
					offset = 0
				case o.Count > l-(l%PRINT_LIMIT):
					offset = l - (l % PRINT_LIMIT)
				default:
					offset = o.Count
				}
			}
			res += fmt.Sprintf("  \u2195 результаты с *%d* по *%d* из *%d*\n", offset+1, offset+cnt, l)
			pages = append(pages, TPagination{OFFSET_SUBNET6, l})
		}

		break
	}

	res += printUpToDate(oldest)

	return
}

func constructResult(a []*pb.Content, o TPagination) (res string, pages []TPagination) {
	var (
		mass   string
		oldest int64 = MAX_TIMESTAMP
		ra     []TReason
	)

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
		ra[0].Ip = append(ra[0].Ip, int2Ip4(a[0].Ip4))
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
				ra[i].Ip = append(ra[i].Ip, int2Ip4(a[i+1].Ip4))
			}
			if len(a[i+1].Ip6) != 0 {
				ra[i].Ip = append(ra[i].Ip, net.IP(a[i+1].Ip6).String())
			}
			if a[i+1].Domain != "" {
				ra[i].Domain = append(ra[i].Domain, PrintedDomain(a[i+1].Domain))
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
				ra[i+1].Ip = append(ra[i+1].Ip, int2Ip4(a[i+1].Ip4))
			}
			if len(a[i+1].Ip6) != 0 {
				ip6 := net.IP(a[i+1].Ip6)
				ra[i+1].Ip = append(ra[i+1].Ip, ip6.String())
			}
			if a[i+1].Domain != "" {
				ra[i+1].Domain = append(ra[i+1].Domain, PrintedDomain(a[i+1].Domain))
			}
			if a[i+1].Url != "" {
				ra[i+1].Url = append(ra[i+1].Url, a[i+1].Url)
			}

		}
	}

	sort.Slice(a, func(j, i int) bool {
		switch {
		case a[i].BlockType == TBLOCK_URL && a[j].BlockType != TBLOCK_URL:
			return true
		case a[i].BlockType == TBLOCK_HTTPS &&
			(a[j].BlockType != TBLOCK_URL &&
				a[j].BlockType != TBLOCK_HTTPS):
			return true
		case a[i].BlockType == TBLOCK_DOMAIN &&
			(a[j].BlockType != TBLOCK_URL &&
				a[j].BlockType != TBLOCK_HTTPS &&
				a[j].BlockType != TBLOCK_DOMAIN):
			return true
		case a[i].BlockType == TBLOCK_MASK &&
			(a[j].BlockType != TBLOCK_URL &&
				a[j].BlockType != TBLOCK_HTTPS &&
				a[j].BlockType != TBLOCK_DOMAIN &&
				a[j].BlockType != TBLOCK_MASK):
			return true
		default:
			return false
		}
	})

	offset := 0
	if o.Tag == OFFSET_CONTENT {
		switch {
		case len(a) <= PRINT_LIMIT:
			offset = 0
		case o.Count > len(a)-(len(a)%PRINT_LIMIT):
			offset = len(a) - (len(a) % PRINT_LIMIT)
		default:
			offset = o.Count
		}
	}

	var cnt, cbu, cbh, cbd, cbm, cbi int
	for i, packet := range a {
		if o.Tag == OFFSET_CONTENT && i < offset {
			continue
		}

		content := Content{}
		err := json.Unmarshal(packet.Pack, &content)
		if err != nil {
			logger.Error.Printf("Упс!!! %s\n", err)
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
				mass = "\U0001f4a5\U0001f4a5\U0001f4a5 Ресурс под «ковровой» блокировкой!!\n\n"
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
			res += fmt.Sprintf("%s /n\\_%d %s /d\\_%s\n", bt, content.ID, dcs, String2fnv2base32(dcs))
			res += fmt.Sprintf("\u2022 %s\n", constructBasis(content.EntryType, printOrg(content.Decision.Org), content.Decision.Number))
			if len(req.Aggr) != 0 {
				for _, nw := range req.Aggr {
					res += fmt.Sprintf("    _как подсеть_ %s\n", nw)
				}
			}
			if len(req.Ip) != 0 {
				for _, ip := range req.Ip {
					res += fmt.Sprintf("    _как ip_ %s\n", ip)
				}
			}
			if len(req.Domain) != 0 {
				for _, domain := range req.Domain {
					res += fmt.Sprintf("    _как domain_ %s\n", Sanitize(PrintedDomain(domain)))
				}
			}
			if len(req.Url) != 0 {
				for _, u := range req.Url {
					res += fmt.Sprintf("    _как url_ %s\n", Sanitize(PrintedDomain(u)))
				}
			}
			res += "\n"

			points := ""
			switch packet.BlockType {
			case TBLOCK_URL:
				l := len(content.URL)
				if l > PRINT_LIMIT {
					l = PRINT_LIMIT
					points = "..."
				}

				urls := make([]string, 0, l)
				for _, u := range content.URL[:l] {
					urls = append(urls, Sanitize(u.URL))
				}

				res += fmt.Sprintf("  urls: %s%s\n", strings.Join(urls, ", "), points)
			case TBLOCK_HTTPS, TBLOCK_DOMAIN, TBLOCK_MASK:
				l := len(content.Domain)
				if l > PRINT_LIMIT {
					l = PRINT_LIMIT
					points = "..."
				}

				domains := make([]string, 0, l)
				for _, d := range content.Domain[:l] {
					domains = append(domains, Sanitize(d.Domain))
				}

				res += fmt.Sprintf("  domains: %s%s\n", strings.Join(domains, ", "), points)
			case TBLOCK_IP:
				l := len(content.IPv4) + len(content.IPv6) + len(content.SubnetIPv4) + len(content.SubnetIPv6)
				if l > PRINT_LIMIT {
					l = PRINT_LIMIT
					points = "..."
				}

				ips := make([]string, 0, l)

				cnt := PRINT_LIMIT
				for _, ip := range content.IPv4 {
					if cnt == 0 {
						break
					}

					ips = append(ips, int2Ip4(ip.IPv4))
					cnt--
				}

				for _, ip := range content.IPv6 {
					if cnt == 0 {
						break
					}

					ips = append(ips, net.IP(ip.IPv6).String())
					cnt--
				}

				for _, sb := range content.SubnetIPv4 {
					if cnt == 0 {
						break
					}

					ips = append(ips, sb.SubnetIPv4)
					cnt--
				}

				for _, sb := range content.SubnetIPv6 {
					if cnt == 0 {
						break
					}

					ips = append(ips, sb.SubnetIPv6)
					cnt--
				}

				res += fmt.Sprintf("  ips: %s%s\n", strings.Join(ips, ", "), points)
			}
		}
		cnt++
	}

	if mass != "" {
		res = mass + res
	}

	if len(a) > PRINT_LIMIT {
		pages = append(pages, TPagination{OFFSET_CONTENT, len(a)})
		// rest := cnt - PRINT_LIMIT
		res += fmt.Sprintf("\u2195 результаты с *%d* по *%d* из *%d*\n", offset+1, offset+PRINT_LIMIT, len(a))
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
	res += "*типы блокировки:* " + strings.Join(abt, " | ")
	res += printUpToDate(oldest)
	return
}
