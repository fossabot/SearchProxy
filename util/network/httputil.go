package network

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"searchproxy/util/miscellaneous"

	log "github.com/sirupsen/logrus"
)

// HTTPHEAD - run HTTP HEAD request against URL
func (hu *HTTPUtilities) HTTPHEAD(url string) (res *http.Response, err error) {
	client := &http.Client{}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), hu.RequestTimeout)
	defer cancel()

	req = req.WithContext(ctx)
	req.Header.Set("User-Agent",
		fmt.Sprintf("Mozilla/5.0 (compatible; SearchProxy/%s; %s; +https://github.com/tb0hdan/SearchProxy)",
			hu.BuildInfo.Version, hu.BuildInfo.GoVersion))

	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// PingHTTP - run HTTP HEAD against URL and measure response time
func (hu *HTTPUtilities) PingHTTP(url string) (elapsed int64) {
	start := time.Now().UnixNano()
	res, err := hu.HTTPHEAD(url)
	elapsed = (time.Now().UnixNano() - start) / time.Millisecond.Nanoseconds()

	if err != nil {
		log.Debugf("An error %v occurred while running ping on %s", err, url)
		// failed servers should be marked as slow, with negative values
		elapsed = MirrorUnreachable * elapsed
	} else {
		res.Body.Close()
	}

	return
}

// NewHTTPUtilities - create new http utilities instance
func NewHTTPUtilities(buildInfo *miscellaneous.BuildInfo, timeout time.Duration) *HTTPUtilities {
	return &HTTPUtilities{
		BuildInfo:      buildInfo,
		RequestTimeout: timeout,
	}
}

// StripRequestURI - remove prefix from URI
func StripRequestURI(requestURI, prefix string) (result string) {
	result = strings.TrimLeft(requestURI, prefix)
	if !strings.HasPrefix(result, "/") {
		result = "/" + result
	}

	return
}

// GetRemoteAddressFromRequest - returns remote address based on request headers. Respects X-Forwarded-For
func GetRemoteAddressFromRequest(r *http.Request) (addr string, err error) {
	var (
		remoteAddr string
	)

	remoteAddr, _, err = net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		// Something's very wrong with the request
		return "", err
	}

	addr = r.Header.Get("X-Real-IP")

	if len(addr) == 0 {
		addr = r.Header.Get("X-Forwarded-For")
	}

	// Could not get IP from headers
	if len(addr) == 0 {
		addr = remoteAddr
	} else if !IsLocalNetworkString(remoteAddr) { // IP is from headers, check whether we can trust it
		// Nope, use remote address instead
		addr = remoteAddr
	}

	return addr, nil
}

// WriteResponse - shorthand function for writing HTTP responses
func WriteResponse(w http.ResponseWriter, statusCode int, content string) {
	w.WriteHeader(statusCode)
	fmt.Fprint(w, content)
}

// WriteNormalResponse - shorthand function for 200 OK replies
func WriteNormalResponse(w http.ResponseWriter, content string) {
	WriteResponse(w, http.StatusOK, content)
}

// WriteNotFound - shorthand function for 404 not found replies
func WriteNotFound(w http.ResponseWriter) {
	WriteResponse(w, http.StatusNotFound, "404 page not found")
}
