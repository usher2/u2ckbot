package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/usher2/u2ckbot/internal/logger"
	pb "github.com/usher2/u2ckbot/msg"
)

const (
	ErrorMessageSomethingGoingWrong = "\U00002620 Что-то пошло не так! Повторите попытку позже"
	ErrorMesssageTryAgainLater      = "\u23f3 Повторите попытку позже: %s"
)

func errMsgTryAgainLater(s string) string {
	return fmt.Sprintf(ErrorMesssageTryAgainLater, s)
}

func printOrg(org string) string {
	if org == "" {
		return "Генпрокуратура"
	}

	return org
}

// Summarize - summarize stats.
func Summarize(c pb.CheckClient) string {
	logger.Info.Printf("Looking for Summarize\n")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := c.Summary(ctx, &pb.SummaryRequest{Query: "summary"})
	if err != nil {
		logger.Debug.Printf("%v.Summarize(_) = _, %v\n", c, err)

		return ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return errMsgTryAgainLater(r.Error)
	}

	return printSummary(r.Summary)
}

func printSummary(s []byte) string {
	var stats SummaryValues

	if err := json.Unmarshal(s, &stats); err != nil {
		logger.Error.Printf("ERROR: %s\n", err.Error())

		return ErrorMessageSomethingGoingWrong
	}

	res := "*Сводная статистика по выгрузке*\n\n"

	res += fmt.Sprintf("*Всего записей:* %d\n", stats.ContentEntries)
	res += fmt.Sprintf("*URL:* %d\n", stats.URLEntries)
	res += fmt.Sprintf("*Доменов:* %d\n", stats.DomainEntries)
	res += fmt.Sprintf("*IPv4-адресов:* %d\n", stats.IPv4Entries)
	res += fmt.Sprintf("*IPv6-адресов:* %d\n", stats.IPv6Entries)
	res += fmt.Sprintf("*IPv4 сетей:* %d\n", stats.SubnetIPv4Entries)
	res += fmt.Sprintf("*IPv6 сетей:* %d\n", stats.SubnetIPv6Entries)

	res += "\n_Записей по типам блокировок:_\n"

	res += fmt.Sprintf("    *%s*: %d\n", BlockTypeView(BlockTypeURL), stats.BlockTypeURL)
	res += fmt.Sprintf("    *%s*: %d\n", BlockTypeView(BlockTypeHTTPS), stats.BlockTypeHTTPS)
	res += fmt.Sprintf("    *%s*: %d\n", BlockTypeView(BlockTypeDomain), stats.BlockTypeDomain)
	res += fmt.Sprintf("    *%s*: %d\n", BlockTypeView(BlockTypeMask), stats.BlockTypeMask)
	res += fmt.Sprintf("    *%s*: %d\n", BlockTypeView(BlockTypeIP), stats.BlockTypeIP)

	res += "\n_Записей по принявшим решения организациям:_\n"

	orgs := make([]string, 0, len(stats.DecisionOrgs))
	for org := range stats.DecisionOrgs {
		orgs = append(orgs, org)
	}

	sort.Slice(orgs, func(i, j int) bool {
		return stats.DecisionOrgs[orgs[i]] > stats.DecisionOrgs[orgs[j]]
	})

	for _, org := range orgs {
		num := stats.DecisionOrgs[org]
		switch {
		case num < 2000:
			res += fmt.Sprintf("    *%s* /o\\_%s : %d\n", org, Sanitize(String2fnv2base32(org)), num)
		default:

			res += fmt.Sprintf("    *%s*: %d\n", org, num)
		}
	}

	res += "\n_Записей по типам решений:_\n"

	types := make([]string, 0, len(stats.EntryTypes))
	for t := range stats.EntryTypes {
		types = append(types, t)
	}

	sort.Slice(types, func(i, j int) bool {
		return stats.EntryTypes[types[i]] > stats.EntryTypes[types[j]]
	})

	for _, t := range types {
		link, desc, _ := strings.Cut(DecisionTypeView(t), ",")
		num := stats.EntryTypes[t]
		switch {
		case num < 2000:
			res += fmt.Sprintf("    %s, *%s* /e\\_%s : %d\n", link, desc, Sanitize(t), num)
		default:
			res += fmt.Sprintf("    %s, *%s*: %d\n", link, desc, num)
		}
	}

	res += fmt.Sprintf("\n*Самая большая запись (%3.1f MB):* /n\\_%d\n", float64(stats.LargestSizeOfContent)/1024/1024, stats.LargestSizeOfContentCintentID)

	res += fmt.Sprintf("*Больше всего ссылок (%d) на объект:* `%s`\n", stats.MaxItemReferences, stats.MaxItemReferencesString)

	res += fmt.Sprintf("*Решений без номера (б/н):* %d\n", stats.EntriesWithoutDecisionNo)

	res += printUpToDate(stats.UpdateTime)

	return res
}

// Ping - ping API server.
func Ping(c pb.CheckClient) string {
	logger.Info.Printf("Looking for Ping\n")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := c.Ping(ctx, &pb.PingRequest{Ping: "ping"})
	if err != nil {
		logger.Debug.Printf("%v.Ping(_) = _, %v\n", c, err)

		return ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return errMsgTryAgainLater(r.Error)
	}

	return fmt.Sprintf("\U0001f919 *%s*%s", r.Pong, printUpToDate(r.RegistryUpdateTime))
}

func searchContentID(c pb.CheckClient, id int) (int64, []*pb.Content, string) {
	logger.Info.Printf("Looking for content: #%d\n", id)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := c.SearchContentID(ctx, &pb.ContentIDRequest{Query: int32(id)})
	if err != nil {
		logger.Debug.Printf("%v.SearchContent(_) = _, %v\n", c, err)

		return MAX_TIMESTAMP, nil, ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return MAX_TIMESTAMP, nil, errMsgTryAgainLater(r.Error)
	}

	return r.RegistryUpdateTime, r.Results[:], ""
}

func searchIPv4(c pb.CheckClient, ip netip.Addr) (int64, []*pb.Content, string) {
	logger.Info.Printf("Looking for %s\n", ip)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := c.SearchIPv4(ctx, &pb.IPv4Request{Query: parseIp4(ip.String())})
	if err != nil {
		logger.Debug.Printf("%v.SearchIP4(_) = _, %v\n", c, err)

		return MAX_TIMESTAMP, nil, ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return MAX_TIMESTAMP, nil, errMsgTryAgainLater(r.Error)
	}

	return r.RegistryUpdateTime, r.Results[:], ""
}

func searchIPv6(c pb.CheckClient, ip netip.Addr) (int64, []*pb.Content, string) {
	logger.Info.Printf("Looking for %s\n", ip)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := c.SearchIPv6(ctx, &pb.IPv6Request{Query: ip.AsSlice()})
	if err != nil {
		logger.Debug.Printf("%v.SearchIP6(_) = _, %v\n", c, err)

		return MAX_TIMESTAMP, nil, ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return MAX_TIMESTAMP, nil, errMsgTryAgainLater(r.Error)
	}

	return r.RegistryUpdateTime, r.Results[:], ""
}

func searchURL(c pb.CheckClient, u string) (int64, []*pb.Content, string) {
	_url := NormalizeURL(u)
	if _url != u {
		fmt.Printf("Input was %s\n", u)
	}

	logger.Info.Printf("Looking for %s\n", _url)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := c.SearchURL(ctx, &pb.URLRequest{Query: _url})
	if err != nil {
		logger.Debug.Printf("%v.SearchURL(_) = _, %v\n", c, err)

		return MAX_TIMESTAMP, nil, ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return MAX_TIMESTAMP, nil, errMsgTryAgainLater(r.Error)
	}

	return r.RegistryUpdateTime, r.Results[:], ""
}

func searchDomain(c pb.CheckClient, s string) (int64, []*pb.Content, string) {
	domain := NormalizeDomain(s)

	logger.Info.Printf("Looking for %s\n", domain)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r, err := c.SearchDomain(ctx, &pb.DomainRequest{Query: domain})
	if err != nil {
		logger.Debug.Printf("%v.SearchDomain(_) = _, %v\n", c, err)

		return MAX_TIMESTAMP, nil, ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return MAX_TIMESTAMP, nil, errMsgTryAgainLater(r.Error)
	}

	return r.RegistryUpdateTime, r.Results[:], ""
}

func searchDomainSuffix(c pb.CheckClient, s string, variant int32) (int64, []*pb.Content, string) {
	domain := NormalizeDomain(s)

	logger.Info.Printf("Looking for suffix %s\n", domain)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r, err := c.SearchDomainSuffix(ctx, &pb.SuffixRequest{Query: domain, Variant: variant})
	if err != nil {
		logger.Debug.Printf("%v.SearchDomainSuffix(_) = _, %v\n", c, err)

		return MAX_TIMESTAMP, nil, ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return MAX_TIMESTAMP, nil, errMsgTryAgainLater(r.Error)
	}

	return r.RegistryUpdateTime, r.Results[:], ""
}

func searchEntryType(c pb.CheckClient, s string) (int64, []*pb.Content, string) {
	logger.Info.Printf("Looking for entry type %s\n", s)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r, err := c.SearchEntryType(ctx, &pb.EntryTypeRequest{Query: s})
	if err != nil {
		logger.Debug.Printf("%v.SearchEntryType(_) = _, %v\n", c, err)

		return MAX_TIMESTAMP, nil, ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return MAX_TIMESTAMP, nil, errMsgTryAgainLater(r.Error)
	}

	return r.RegistryUpdateTime, r.Results[:], ""
}

func searchOrg(c pb.CheckClient, s uint64) (int64, []*pb.Content, string, string) {
	logger.Info.Printf("Looking for org %x\n", s)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r, err := c.SearchOrg(ctx, &pb.OrgRequest{Query: s})
	if err != nil {
		logger.Debug.Printf("%v.SearchOrg(_) = _, %v\n", c, err)

		return MAX_TIMESTAMP, nil, "", ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return MAX_TIMESTAMP, nil, "", errMsgTryAgainLater(r.Error)
	}

	return r.RegistryUpdateTime, r.Results[:], r.Query, ""
}

func searchDecision(c pb.CheckClient, decision uint64) (int64, []*pb.Content, string) {
	logger.Info.Printf("Looking for &%d\n", decision)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r, err := c.SearchDecision(ctx, &pb.DecisionRequest{Query: decision})
	if err != nil {
		logger.Debug.Printf("%v.SearchDecision(_) = _, %v\n", c, err)

		return MAX_TIMESTAMP, nil, ErrorMessageSomethingGoingWrong
	}

	if r.Error != "" {
		logger.Debug.Printf("ERROR: %s\n", r.Error)

		return MAX_TIMESTAMP, nil, errMsgTryAgainLater(r.Error)
	}

	return r.RegistryUpdateTime, r.Results[:], ""
}

// refSearch - searches for domain, IP4 and IP6. Returns oldest timestamp, results and errors.
func refSearch(c pb.CheckClient, s string) (int64, []*pb.Content, []string, []string, string) {
	var (
		oldest int64 = MAX_TIMESTAMP
		a      []*pb.Content
	)

	domain := NormalizeDomain(s)

	ips4 := getIP4(domain)
	for _, ip := range ips4 {
		parsedIP, err := netip.ParseAddr(ip)
		if err != nil {
			continue
		}

		utime, a2, errMsg := searchIPv4(c, parsedIP)
		if errMsg != "" {
			return oldest, nil, nil, nil, errMsg
		}

		if utime < oldest {
			oldest = utime
		}

		a = append(a, a2...)
	}

	ips6 := getIP6(domain)
	for _, ip := range ips6 {
		parsedIP, err := netip.ParseAddr(ip)
		if err != nil {
			continue
		}

		utime, a2, errMsg := searchIPv6(c, parsedIP)
		if errMsg != "" {
			return oldest, nil, nil, nil, errMsg
		}

		if utime < oldest {
			oldest = utime
		}

		a = append(a, a2...)
	}

	return oldest, a, ips4, ips6, ""
}

func domainSuffixSearch(c pb.CheckClient, s string, o TPagination, variant int32) (string, []TPagination) {
	var oldestRecordTimestamp int64 = MAX_TIMESTAMP

	if len(s) == 0 {
		return "\U0001f914 Что имелось ввиду?..\n", nil
	}

	recordUpdateTimestamp, a, errMsg := searchDomainSuffix(c, s, variant)
	if errMsg != "" {
		return errMsg + "\n", nil
	}

	if recordUpdateTimestamp < oldestRecordTimestamp {
		oldestRecordTimestamp = recordUpdateTimestamp
	}

	if len(a) == 0 {
		return fmt.Sprintf("\U0001f914 %s *не найден*\n%s", s, printUpToDate(oldestRecordTimestamp)), nil
	}

	res, pages := constructResult(a, o)

	pattern := SuffixBlockedPattern
	if variant == 2 {
		pattern = SuffixBlockedPatternPlus
	}

	return fmt.Sprintf("\U0001f525 %s*%s*\n\n%s\n", Sanitize(s), pattern, res), pages
}

func entryTypeSearch(c pb.CheckClient, s string, o TPagination) (string, []TPagination) {
	var oldestRecordTimestamp int64 = MAX_TIMESTAMP

	if len(s) == 0 {
		return "\U0001f914 Что имелось ввиду?..\n", nil
	}

	recordUpdateTimestamp, a, errMsg := searchEntryType(c, s)
	if errMsg != "" {
		return errMsg + "\n", nil
	}

	if recordUpdateTimestamp < oldestRecordTimestamp {
		oldestRecordTimestamp = recordUpdateTimestamp
	}

	if len(a) == 0 {
		return fmt.Sprintf("\U0001f914 %s *не найден*\n%s", s, printUpToDate(oldestRecordTimestamp)), nil
	}

	res, pages := constructResult(a, o)

	return fmt.Sprintf("\U0001f4dc /e\\_%s | %s\n\n%s\n", Sanitize(s), DecisionTypeView(s), res), pages
}

func orgSearch(c pb.CheckClient, s string, o TPagination) (string, []TPagination) {
	var oldestRecordTimestamp int64 = MAX_TIMESTAMP

	if len(s) == 0 {
		return "\U0001f914 Что имелось ввиду?..\n", nil
	}

	query, err := Base32ToUint64(s)
	if err != nil {
		return "\U0001f914 Что имелось ввиду?..\n", nil
	}

	recordUpdateTimestamp, a, org, errMsg := searchOrg(c, query)
	if errMsg != "" {
		return errMsg + "\n", nil
	}

	if recordUpdateTimestamp < oldestRecordTimestamp {
		oldestRecordTimestamp = recordUpdateTimestamp
	}

	if len(a) == 0 {
		return fmt.Sprintf("\U0001f914 %s *не найден*\n%s", s, printUpToDate(oldestRecordTimestamp)), nil
	}

	res, pages := constructResult(a, o)

	return fmt.Sprintf("\U0001f4dc *%s* /o\\_%s\n\n%s\n", Sanitize(org), Sanitize(s), res), pages
}

// numberSearch - searches for a internal Roscomnadzor record number.
func numberSearch(c pb.CheckClient, s string, o TPagination) (string, []TPagination) {
	var oldestRecordTimestamp int64 = MAX_TIMESTAMP

	if len(s) == 0 {
		return "\U0001f914 Что имелось ввиду?..\n", nil
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Sprintf("\U0001f914 Что имелось ввиду?.. /n\\_%s: %s\n", s, err.Error()), nil
	}

	if n == 0 {
		return fmt.Sprintf("\U0001f914 Что имелось ввиду?.. /n\\_%s\n", s), nil
	}

	recordUpdateTimestamp, a, errMsg := searchContentID(c, n)
	if errMsg != "" {
		return errMsg + "\n", nil
	}

	if recordUpdateTimestamp < oldestRecordTimestamp {
		oldestRecordTimestamp = recordUpdateTimestamp
	}

	if len(a) == 0 {
		return fmt.Sprintf("\U0001f914 %s *не найден*\n%s", s, printUpToDate(oldestRecordTimestamp)), nil
	}

	return constructContentResult(a, o)
}

func decisionSearch(c pb.CheckClient, s string, o TPagination) (string, []TPagination) {
	var (
		a      []*pb.Content
		oldest int64 = MAX_TIMESTAMP
	)

	if len(s) == 0 {
		return "\U0001f914 Что имелось ввиду?..\n", nil
	}

	n, err := Base32ToUint64(s)
	if err != nil {
		return fmt.Sprintf("\U0001f914 Что имелось ввиду?.. /n\\_%s: %s\n", s, err.Error()), nil
	}

	if n == 0 {
		return fmt.Sprintf("\U0001f914 Что имелось ввиду?.. /n\\_%s\n", s), nil
	}

	utime, a, errMsg := searchDecision(c, n)
	if errMsg != "" {
		return errMsg + "\n", nil
	}

	if utime < oldest {
		oldest = utime
	}

	if len(a) == 0 {
		return fmt.Sprintf("\U0001f914 %s *не найден*\n%s", s, printUpToDate(oldest)), nil
	}

	res, pages := constructResult(a, o)

	content := Content{}
	if err := json.Unmarshal(a[0].Pack, &content); err != nil {
		return fmt.Sprintf("\U0001f914 Что имелось ввиду?.. /n\\_%s: %s\n", s, err.Error()), nil
	}

	dcs := fmt.Sprintf("%s %s %s", content.Decision.Org, content.Decision.Number, content.Decision.Date)

	return fmt.Sprintf("\U0001f4dc /d\\_%s %s\n\n%s", s, dcs, res), pages
}

func mainSearch(c pb.CheckClient, s string, o TPagination) (string, []TPagination) {
	var oldest int64 = MAX_TIMESTAMP

	fmt.Fprintf(os.Stderr, "**** mainSearch: %s\n", s)

	if len(s) == 0 {
		return "\U0001f914 Что имелось ввиду?..\n", nil
	}

	// normalize domain.
	domain := NormalizeDomain(s)

	parsedURL, errParseURL := url.Parse(NormalizeURL(s))
	if errParseURL == nil && parsedURL.IsAbs() &&
		(parsedURL.Scheme == "http" || parsedURL.Scheme == "https") &&
		(parsedURL.Port() == "80" || parsedURL.Port() == "443" || parsedURL.Port() == "") &&
		(parsedURL.RequestURI() == "" || parsedURL.RequestURI() == "/") {
		s = parsedURL.Hostname()
		domain = NormalizeDomain(s)
		errParseURL = fmt.Errorf("fake")
	}

	// try to parse as IP.
	parsedIP, errParseAddr := netip.ParseAddr(s)

	fmt.Fprintf(os.Stderr, "**** mainSearch: %s, %s, %s\n", s, domain, parsedIP)

	switch {
	case errParseAddr == nil:
		fmt.Fprintln(os.Stderr, "**** mainSearch: IP")

		switch {
		case parsedIP.Is4():
			utime, a, errMsg := searchIPv4(c, parsedIP)
			if errMsg != "" {
				return errMsg + "\n", nil
			}

			if utime < oldest {
				oldest = utime
			}

			utime, a2, errMsg := searchDomain(c, s)
			if errMsg != "" {
				return errMsg + "\n", nil
			}

			if len(a2) > 0 {
				a = append(a, a2...)

				if utime < oldest {
					oldest = utime
				}
			}

			if len(a) == 0 {
				return fmt.Sprintf("\U0001f914 %s *не найден*\n%s", Sanitize(s), printUpToDate(oldest)), nil
			}

			res, pages := constructResult(a, o)

			return fmt.Sprintf("\U0001f525 %s *заблокирован*\n\n%s\n", Sanitize(s), res), pages
		case parsedIP.Is6():
			utime, a, errMsg := searchIPv6(c, parsedIP)
			if errMsg != "" {
				return errMsg + "\n", nil
			}

			if utime < oldest {
				oldest = utime
			}

			if len(a) == 0 {
				return fmt.Sprintf("\U0001f914 %s *не найден*\n%s", Sanitize(s), printUpToDate(oldest)), nil
			}

			res, pages := constructResult(a, o)

			return fmt.Sprintf("\U0001f525 %s *заблокирован*\n\n%s\n", Sanitize(s), res), pages

		default:
			return fmt.Sprintf("\U0001f914 Что имелось ввиду?.. %s\n", s), nil
		}
	case isDomainName(domain):
		fmt.Fprintln(os.Stderr, "**** mainSearch: domain")

		utime, a, errMsg := searchDomain(c, domain)
		if errMsg != "" {
			return errMsg + "\n", nil
		}

		if utime < oldest {
			oldest = utime
		}

		switch {
		case strings.HasPrefix(s, "www."):
			utime, a2, errMsg := searchDomain(c, domain[4:])
			if errMsg != "" {
				return errMsg + "\n", nil
			}

			if len(a2) > 0 {
				a = append(a, a2...)

				if utime < oldest {
					oldest = utime
				}
			}
		default:
			utime, a2, errMsg := searchDomain(c, "www."+domain)
			if errMsg != "" {
				return errMsg + "\n", nil
			}

			if len(a2) > 0 {
				a = append(a, a2...)

				if utime < oldest {
					oldest = utime
				}
			}
		}

		if len(a) > 0 {
			res, pages := constructResult(a, o)

			return fmt.Sprintf("\U0001f525 %s *заблокирован*\n\n%s\n", Sanitize(s), res), pages
		}

		text := fmt.Sprintf("\u2705 %s *не заблокирован*\n", Sanitize(s))

		utime, a2, errMsg := searchDomainSuffix(c, domain, 2)
		if errMsg != "" {
			return errMsg + "\n", nil
		}

		if len(a2) > 0 {
			a = append(a, a2...)

			if utime < oldest {
				oldest = utime
			}

			text += "\n\U0001f525 но может быть ограничен по базовому домену:\n"
		}

		utime, a2, ips4, ips6, errMsg := refSearch(c, domain)
		if errMsg != "" {
			return errMsg + "\n", nil
		}

		if len(a2) > 0 {
			a = append(a, a2...)

			if utime < oldest {
				oldest = utime
			}

			text += "\n\U0001f525 но может быть ограничен по IP-адресу:\n"
			for _, ip := range ips4 {
				text += fmt.Sprintf("    %s\n", ip)
			}
			for _, ip := range ips6 {
				text += fmt.Sprintf("    %s\n", ip)
			}
		}

		if len(a) == 0 {
			return text + printUpToDate(oldest), nil
		}

		res, pages := constructResult(a, o)

		return fmt.Sprintf("%s\n%s\n", text, res), pages
	case errParseURL == nil:
		fmt.Fprintln(os.Stderr, "**** mainSearch: URL")

		utime, a, errMsg := searchURL(c, parsedURL.String())
		if errMsg != "" {
			return errMsg + "\n", nil
		}

		if utime < oldest {
			oldest = utime
		}

		switch parsedURL.Scheme {
		case "http":
			parsedURL.Scheme = "https"
			utime, a2, errMsg := searchURL(c, parsedURL.String())
			if errMsg != "" {
				return errMsg + "\n", nil
			}

			if len(a2) > 0 {
				a = append(a, a2...)

				if utime < oldest {
					oldest = utime
				}
			}
		case "https":
			parsedURL.Scheme = "http"
			utime, a2, errMsg := searchURL(c, parsedURL.String())
			if errMsg != "" {
				return errMsg + "\n", nil
			}

			if len(a2) > 0 {
				a = append(a, a2...)

				if utime < oldest {
					oldest = utime
				}
			}

		}

		if len(a) == 0 {
			return fmt.Sprintf("\u2705 URL %s *не заблокирован*\n%s", Sanitize(s), printUpToDate(oldest)), nil
		}

		res, pages := constructResult(a, o)

		return fmt.Sprintf("\U0001f525 URL %s *заблокирован*\n\n%s\n", Sanitize(s), res), pages

	default:
		fmt.Fprintln(os.Stderr, "**** mainSearch: default")

		utime, a, errMsg := searchURL(c, s)
		if errMsg != "" {
			return errMsg + "\n", nil
		}

		if utime < oldest {
			oldest = utime
		}

		utime, a2, errMsg := searchDomain(c, s)
		if errMsg != "" {
			return errMsg + "\n", nil
		}

		if len(a2) > 0 {
			a = append(a, a2...)

			if utime < oldest {
				oldest = utime
			}
		}

		if len(a) == 0 {
			return fmt.Sprintf("\U0001f914 Что имелось ввиду?.. %s\n%s", s, printUpToDate(oldest)), nil
		}

		res, pages := constructResult(a, o)

		return fmt.Sprintf("\U0001f525 %s *заблокирован*\n\n%s\n", Sanitize(s), res), pages
	}
}
