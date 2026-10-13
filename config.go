package fystack

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

const apiPrefix = "/api/v1"

func baseURL(environment Environment, domain string) (string, error) {
	if domain != "" {
		if err := validateDomain(domain); err != nil {
			return "", err
		}
		return "https://" + domain + apiPrefix, nil
	}

	switch environment {
	case "", EnvironmentProduction:
		return "https://api.fystack.io" + apiPrefix, nil
	case EnvironmentSandbox:
		return "https://api.enterprise-sandbox.fystack.io" + apiPrefix, nil
	case EnvironmentLocal:
		return "http://localhost:8150" + apiPrefix, nil
	default:
		return "", fmt.Errorf("invalid environment %q", environment)
	}
}

func validateDomain(domain string) error {
	if strings.ContainsAny(domain, "/:@") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Errorf("invalid domain %q: expected a domain like api.example.com", domain)
	}
	if net.ParseIP(domain) != nil {
		return fmt.Errorf("invalid domain %q: IP addresses are not supported", domain)
	}
	parsed, err := url.Parse("https://" + domain)
	if err != nil || parsed.Hostname() != domain || !strings.Contains(domain, ".") {
		return fmt.Errorf("invalid domain %q: expected a domain like api.example.com", domain)
	}
	for _, label := range strings.Split(domain, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("invalid domain %q: expected a domain like api.example.com", domain)
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' {
				return fmt.Errorf("invalid domain %q: expected a domain like api.example.com", domain)
			}
		}
	}
	tld := domain[strings.LastIndexByte(domain, '.')+1:]
	if len(tld) < 2 {
		return fmt.Errorf("invalid domain %q: expected a domain like api.example.com", domain)
	}
	for _, char := range tld {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') {
			return fmt.Errorf("invalid domain %q: expected a domain like api.example.com", domain)
		}
	}
	return nil
}
