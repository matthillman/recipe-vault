package recipeimport

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxHTMLBytes  = 5 << 20
	maxMediaBytes = 10 << 20
)

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

func FetchSource(client HTTPDoer, sourceURL string) (FetchedSource, error) {
	normalized, err := NormalizeSourceURL(sourceURL)
	if err != nil {
		return FetchedSource{}, err
	}
	if client == nil {
		client = SecureHTTPClient()
	}
	req, err := http.NewRequest(http.MethodGet, normalized, nil)
	if err != nil {
		return FetchedSource{}, err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/pdf,image/png,image/jpeg,image/webp,image/gif")
	req.Header.Set("User-Agent", "recipe-vault-importer/1.0 (+personal recipe import)")
	resp, err := client.Do(req)
	if err != nil {
		return FetchedSource{}, fmt.Errorf("fetch source: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return FetchedSource{}, fmt.Errorf("fetch source: HTTP %d", resp.StatusCode)
	}
	declaredType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	declaredType = canonicalMediaType(declaredType)
	if declaredType != "" && declaredType != "application/octet-stream" && !supportedMediaType(declaredType) {
		return FetchedSource{}, fmt.Errorf("unsupported source content type %q", declaredType)
	}
	if resp.ContentLength > maxMediaBytes {
		return FetchedSource{}, fmt.Errorf("source exceeds %d bytes", maxMediaBytes)
	}
	limited := io.LimitReader(resp.Body, maxMediaBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return FetchedSource{}, fmt.Errorf("read source: %w", err)
	}
	if len(body) > maxMediaBytes {
		return FetchedSource{}, fmt.Errorf("source exceeds %d bytes", maxMediaBytes)
	}
	detectedType, _, _ := mime.ParseMediaType(http.DetectContentType(body))
	detectedType = canonicalMediaType(detectedType)
	mediaType := declaredType
	if mediaType == "" || mediaType == "application/octet-stream" {
		mediaType = detectedType
	}
	if !supportedMediaType(mediaType) {
		return FetchedSource{}, fmt.Errorf("unsupported source content type %q", mediaType)
	}
	if mediaType == "application/pdf" && detectedType != "application/pdf" {
		return FetchedSource{}, fmt.Errorf("source declared as PDF but content is %q", detectedType)
	}
	if isImageMediaType(mediaType) && detectedType != mediaType {
		return FetchedSource{}, fmt.Errorf("source declared as %s but content is %q", mediaType, detectedType)
	}
	if isHTMLMediaType(mediaType) && supportedMediaType(detectedType) && !isHTMLMediaType(detectedType) {
		return FetchedSource{}, fmt.Errorf("source declared as HTML but content is %q", detectedType)
	}
	if isHTMLMediaType(mediaType) && len(body) > maxHTMLBytes {
		return FetchedSource{}, fmt.Errorf("HTML source exceeds %d bytes", maxHTMLBytes)
	}
	finalRequest := resp.Request
	if finalRequest == nil {
		finalRequest = req
	}
	finalURL, err := NormalizeSourceURL(finalRequest.URL.String())
	if err != nil {
		return FetchedSource{}, err
	}
	return FetchedSource{Body: body, URL: finalURL, MediaType: mediaType}, nil
}

func isHTMLMediaType(mediaType string) bool {
	return mediaType == "text/html" || mediaType == "application/xhtml+xml"
}

func supportedMediaType(mediaType string) bool {
	return isHTMLMediaType(mediaType) || mediaType == "application/pdf" || isImageMediaType(mediaType)
}

func isImageMediaType(mediaType string) bool {
	return mediaType == "image/png" || mediaType == "image/jpeg" ||
		mediaType == "image/webp" || mediaType == "image/gif"
}

func canonicalMediaType(mediaType string) string {
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	if mediaType == "image/jpg" {
		return "image/jpeg"
	}
	return mediaType
}
