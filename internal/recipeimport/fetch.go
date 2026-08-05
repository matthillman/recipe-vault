package recipeimport

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxPageBytes = 5 << 20

func NormalizeSourceURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("parse source URL: %w", err)
	}
	if u.Scheme != "https" {
		return "", fmt.Errorf("source URL must use https")
	}
	if u.Hostname() == "" || strings.EqualFold(u.Hostname(), "localhost") {
		return "", fmt.Errorf("source URL must have a public hostname")
	}
	u.Fragment = ""
	query := u.Query()
	for key := range query {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "utm_") || lower == "fbclid" || lower == "gclid" {
			query.Del(key)
		}
	}
	u.RawQuery = query.Encode()
	u.Host = strings.ToLower(u.Host)
	return u.String(), nil
}

func SecureHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}
			for _, ip := range ips {
				if !publicIP(ip) {
					continue
				}
				conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
				if dialErr == nil {
					return conn, nil
				}
				err = dialErr
			}
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("hostname %q resolves only to non-public addresses", host)
		},
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		IdleConnTimeout:       30 * time.Second,
	}
	client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		_, err := NormalizeSourceURL(req.URL.String())
		return err
	}
	return client
}

func publicIP(ip net.IP) bool {
	return ip != nil && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsUnspecified() &&
		!ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsMulticast()
}

func FetchPage(client HTTPDoer, sourceURL string) (string, string, error) {
	normalized, err := NormalizeSourceURL(sourceURL)
	if err != nil {
		return "", "", err
	}
	if client == nil {
		client = SecureHTTPClient()
	}
	req, err := http.NewRequest(http.MethodGet, normalized, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("User-Agent", "recipe-vault-importer/1.0 (+personal recipe import)")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("fetch source: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("fetch source: HTTP %d", resp.StatusCode)
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if contentType != "" && !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml+xml") {
		return "", "", fmt.Errorf("source is not HTML (%s)", contentType)
	}
	limited := io.LimitReader(resp.Body, maxPageBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", "", fmt.Errorf("read source: %w", err)
	}
	if len(body) > maxPageBytes {
		return "", "", fmt.Errorf("source exceeds %d bytes", maxPageBytes)
	}
	finalURL, err := NormalizeSourceURL(resp.Request.URL.String())
	if err != nil {
		return "", "", err
	}
	return string(body), finalURL, nil
}
