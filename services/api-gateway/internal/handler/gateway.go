package handler

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/api-gateway/internal/config"
)

// Gateway reverse-proxies /api/* to backend services (team-guide routing).
type Gateway struct {
	routes    []route
	log       *logger.Logger
	transport http.RoundTripper
}

type route struct {
	apiPrefix     string
	target        *url.URL
	backendPrefix string
	proxy         *httputil.ReverseProxy
}

// NewGateway builds proxy routes from config. Required upstreams: auth, doctor, ai-symptom.
func NewGateway(cfg *config.Config, log *logger.Logger) (*Gateway, error) {
	t := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 120 * time.Second,
	}

	authURL, err := url.Parse(cfg.AuthServiceURL)
	if err != nil {
		return nil, fmt.Errorf("auth service url: %w", err)
	}
	doctorURL, err := url.Parse(cfg.DoctorServiceURL)
	if err != nil {
		return nil, fmt.Errorf("doctor service url: %w", err)
	}
	aiURL, err := url.Parse(cfg.AISymptomServiceURL)
	if err != nil {
		return nil, fmt.Errorf("ai symptom service url: %w", err)
	}
	patientURL, err := url.Parse(cfg.PatientServiceURL)
	if err != nil {
		return nil, fmt.Errorf("patient service url: %w", err)
	}
	appointmentURL, err := url.Parse(cfg.AppointmentServiceURL)
	if err != nil {
		return nil, fmt.Errorf("appointment service url: %w", err)
	}
	paymentURL, err := url.Parse(cfg.PaymentServiceURL)
	if err != nil {
		return nil, fmt.Errorf("payment service url: %w", err)
	}
	telemedicineURL, err := url.Parse(cfg.TelemedicineServiceURL)
	if err != nil {
		return nil, fmt.Errorf("telemedicine service url: %w", err)
	}
	fileStorageURL, err := url.Parse(cfg.FileStorageServiceURL)
	if err != nil {
		return nil, fmt.Errorf("file storage service url: %w", err)
	}

	defs := []struct {
		apiPrefix     string
		rawBase       string
		backendPrefix string
	}{
		{"/api/notifications", cfg.NotificationServiceURL, "/notifications"},
		{"/api/v1/appointments", cfg.AppointmentServiceURL, "/api/v1/appointments"},
		{"/api/v1/slots", cfg.AppointmentServiceURL, "/api/v1/slots"},
		{"/api/ai/symptom", cfg.AISymptomServiceURL, "/symptoms"},
		{"/api/v1/telemedicine", cfg.TelemedicineServiceURL, "/api/v1/telemedicine"},
		{"/api/v1/patient", cfg.PatientServiceURL, "/api/v1/patient"},
		{"/api/v1/payments", cfg.PaymentServiceURL, "/api/v1/payments"},
		{"/api/v1/doctors", cfg.DoctorServiceURL, "/doctors"},
		{"/api/doctors", cfg.DoctorServiceURL, "/doctors"},
		{"/api/admin", cfg.AdminServiceURL, "/admin"},
		{"/api/v1/files", cfg.FileStorageServiceURL, "/api/v1/files"},
		{"/api/auth", cfg.AuthServiceURL, "/auth"},
		{"/api/support", cfg.SupportServiceURL, "/api/support"},
	}

	var routes []route
	for _, d := range defs {
		if strings.TrimSpace(d.rawBase) == "" {
			continue
		}
		var target *url.URL
		switch d.apiPrefix {
		case "/api/auth":
			target = authURL
		case "/api/v1/doctors", "/api/doctors":
			target = doctorURL
		case "/api/ai/symptom":
			target = aiURL
		case "/api/v1/patient":
			target = patientURL
		case "/api/v1/appointments", "/api/v1/slots":
			target = appointmentURL
		case "/api/v1/payments":
			target = paymentURL
		case "/api/v1/telemedicine":
			target = telemedicineURL
		case "/api/v1/files":
			target = fileStorageURL
		default:
			u, err := url.Parse(d.rawBase)
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", d.apiPrefix, err)
			}
			target = u
		}
		routes = append(routes, route{
			apiPrefix:     d.apiPrefix,
			target:        target,
			backendPrefix: d.backendPrefix,
			proxy:         newReverseProxy(target, d.apiPrefix, d.backendPrefix, t, log),
		})
	}

	sort.Slice(routes, func(i, j int) bool {
		return len(routes[i].apiPrefix) > len(routes[j].apiPrefix)
	})

	apiPrefixes := make([]string, len(routes))
	for i, r := range routes {
		apiPrefixes[i] = r.apiPrefix
	}
	log.Info("API gateway route prefixes", "count", len(routes), "prefixes", strings.Join(apiPrefixes, ", "))

	return &Gateway{routes: routes, log: log, transport: t}, nil
}

func newReverseProxy(target *url.URL, apiPrefix, backendPrefix string, rt http.RoundTripper, log *logger.Logger) *httputil.ReverseProxy {
	p := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			appendForwardHeaders(pr)
			path := pr.In.URL.Path
			rest := ""
			if strings.HasPrefix(path, apiPrefix) {
				rest = path[len(apiPrefix):]
			}
			if rest != "" && !strings.HasPrefix(rest, "/") {
				rest = "/" + rest
			}
			if rest == "" {
				pr.Out.URL.Path = backendPrefix
			} else {
				pr.Out.URL.Path = backendPrefix + rest
			}
			pr.Out.URL.RawQuery = pr.In.URL.RawQuery
			if _, ok := pr.Out.Header["User-Agent"]; !ok {
				pr.Out.Header.Set("User-Agent", "")
			}
		},
		Transport: rt,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error("proxy error", "path", r.URL.Path, "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"success":false,"error":"upstream service unavailable"}`))
		},
	}
	return p
}

func appendForwardHeaders(pr *httputil.ProxyRequest) {
	host, _, err := net.SplitHostPort(pr.In.RemoteAddr)
	if err != nil {
		host = pr.In.RemoteAddr
	}
	pr.Out.Header.Set("X-Real-IP", host)
	if prior := pr.In.Header.Get("X-Forwarded-For"); prior == "" {
		pr.Out.Header.Set("X-Forwarded-For", host)
	} else {
		pr.Out.Header.Set("X-Forwarded-For", prior+", "+host)
	}
	if pr.In.TLS != nil {
		pr.Out.Header.Set("X-Forwarded-Proto", "https")
	} else {
		pr.Out.Header.Set("X-Forwarded-Proto", "http")
	}
}

func matchesPrefix(path, prefix string) bool {
	if len(path) < len(prefix) {
		return false
	}
	if path[:len(prefix)] != prefix {
		return false
	}
	if len(path) == len(prefix) {
		return true
	}
	switch path[len(prefix)] {
	case '/', '?':
		return true
	default:
		return false
	}
}

// Handle proxies /api/* or returns JSON errors.
func (g *Gateway) Handle(c *gin.Context) {
	path := c.Request.URL.Path

	if !strings.HasPrefix(path, "/api/") {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "route not found"})
		return
	}

	for _, r := range g.routes {
		if matchesPrefix(path, r.apiPrefix) {
			r.proxy.ServeHTTP(c.Writer, c.Request)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "no upstream configured for this API path",
	})
}
