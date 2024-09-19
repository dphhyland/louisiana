package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)

	logger := slog.NewLogLogger(handler, slog.LevelError)
	log := slog.New(handler)
	slog.SetDefault(log)

	// Load environment variables with fallbacks
	caCertPath := getEnv("CA_CERT_PATH", "certs/ca.crt")
	serverCertPath := getEnv("SERVER_CERT_PATH", "certs/mtls.crt")
	serverKeyPath := getEnv("SERVER_KEY_PATH", "certs/mtls.key")
	authHostStr := getEnv("AUTH_HOST", "http://auth.localhost:3000")
	apiHostStr := getEnv("API_HOST", "http://localhost:8080")

	caBytes, err := os.ReadFile(caCertPath)
	if err != nil {
		log.Error("unable to read ca.crt", slog.String("err", err.Error()))
		os.Exit(1)
	}
	caCertPool := x509.NewCertPool()
	for block, rest := pem.Decode(caBytes); block != nil; block, rest = pem.Decode(rest) {
		switch block.Type {
		case "CERTIFICATE":
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				panic(err)
			}
			caCertPool.AddCert(cert)
			log.Info("loaded certificate", slog.String("subject", cert.Subject.String()))

		default:
			panic("unknown block type " + block.Type)
		}
	}

	serverCertBytes, err := os.ReadFile(serverCertPath)
	if err != nil {
		log.Error("unable to read mtls.crt", slog.String("err", err.Error()))
		os.Exit(1)
	}
	matlsBlock, _ := pem.Decode(serverCertBytes)
	if matlsBlock == nil {
		log.Error("unable to decode mtls.crt")
		os.Exit(1)
	}
	serverBytes := matlsBlock.Bytes

	serverKeyBytes, err := os.ReadFile(serverKeyPath)
	if err != nil {
		log.Error("unable to read mtls.key", slog.String("err", err.Error()))
		os.Exit(1)
	}
	serverKeyBlock, _ := pem.Decode(serverKeyBytes)
	if serverKeyBlock == nil {
		log.Error("unable to decode mtls.key")
		os.Exit(1)
	}
	serverKey, err := x509.ParsePKCS8PrivateKey(serverKeyBlock.Bytes)
	if err != nil {
		log.Error("unable to parse mtls.key", slog.String("err", err.Error()))
		os.Exit(1)
	}

	go http.ListenAndServe(":8181", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>It works!</h1></body></html>"))
	}))

	mux := http.NewServeMux()
	// Auth Host
	authHost, err := url.Parse(authHostStr)
	if err != nil {
		log.Error("unable to upstream url", slog.String("err", err.Error()))
		os.Exit(1)
	}
	// Create reverse proxies with custom director to add headers
	authProxy := httputil.NewSingleHostReverseProxy(authHost)

	// Add custom Director to set required headers
	authProxy.Director = func(req *http.Request) {
		setCustomHeaders(req, authHost)
	}

	// Ensure the externalAuthHost has a trailing slash
	externalAuthHost := getEnv("EXTERNAL_AUTH_HOST", "auth.localhost")
	if !strings.HasSuffix(externalAuthHost, "/") {
		externalAuthHost = externalAuthHost + "/"
	}

	// Use path.Join to ensure safe joining
	mux.Handle(externalAuthHost, accessLogger(log, authProxy))

	// Api Host
	// apiHost, _ := url.Parse(os.Getenv("API_GATEWAY_URI"))
	apiHost, _ := url.Parse(apiHostStr)
	if err != nil {
		log.Error("unable to upstream url", slog.String("err", err.Error()))
		os.Exit(1)
	}
	// Create reverse proxies with custom director to add headers
	apiProxy := httputil.NewSingleHostReverseProxy(apiHost)

	// Add custom Director to set required headers
	apiProxy.Director = func(req *http.Request) {
		setCustomHeaders(req, apiHost)
	}
	apiProxyHandler := accessLogger(log, apiProxy)
	apiProxyHandler = enforceAccessTokenMiddleware(apiProxyHandler)

	// Ensure the externalAuthHost has a trailing slash
	internalApiHost := getEnv("INTERNAL_AUTH_HOST", "api.localhost")
	if !strings.HasSuffix(internalApiHost, "/") {
		internalApiHost = internalApiHost + "/"
	}

	mux.Handle(internalApiHost, apiProxyHandler)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{
				serverBytes,
			},
			PrivateKey: serverKey,
		}},
		InsecureSkipVerify: true,
		ClientCAs:          caCertPool,
		ClientAuth:         tls.VerifyClientCertIfGiven,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		CurvePreferences:   []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},

		CipherSuites: []uint16{
			//TLS 1.2
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			//TLS 1.3 these are actually ignored, but kept here to provide clarity on what's enabled by default.
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
		},
	}

	server := http.Server{
		Handler:   mux,
		ErrorLog:  logger,
		TLSConfig: tlsConfig,
	}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", 443))
	if err != nil {
		os.Exit(1)
	}
	slog.Info("Listening on port 443")
	slog.Error("server error", slog.String("err", server.Serve(tls.NewListener(ln, tlsConfig)).Error()))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func singleJoiningSlash(a, b string) string {
	if a == "" || b == "" {
		return a + b
	}
	aslash := a[len(a)-1] == '/'
	bslash := b[0] == '/'
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func setCustomHeaders(req *http.Request, target *url.URL) {
	req.Header.Set("X-Forwarded-Proto", "https") // Adjust to "https" if using HTTPS
	req.Header.Set("Host", req.Host)
	req.Header.Set("X-Real-IP", getRemoteIP(req))
	req.Header.Set("X-Forwarded-For", getForwardedFor(req))

	// Extract and set the client's certificate and DN
	if len(req.TLS.PeerCertificates) > 0 {
		// The TLS Block Ensures that the correct ordering has taken place and that the leaf certificate will be at block 0
		clientCert := req.TLS.PeerCertificates[0]
		certPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: clientCert.Raw,
		})
		// Base64 encode the certificate to ensure it's valid for HTTP headers
		certPEMString := strings.ReplaceAll(string(certPEM), "\n", " ")
		req.Header.Set("X-Client-Cert", string(certPEMString))
		req.Header.Set("X-Client-DN", clientCert.Subject.String())
	}

	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	if target.RawQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = target.RawQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = target.RawQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
}

func getRemoteIP(req *http.Request) string {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return ""
	}
	return ip
}

func getForwardedFor(req *http.Request) string {
	forwardedFor := req.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		return forwardedFor + ", " + getRemoteIP(req)
	}
	return getRemoteIP(req)
}

func accessLogger(log *slog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}
		h.ServeHTTP(rec, r)
		attrs := []slog.Attr{
			slog.String("remoteIP", r.RemoteAddr),
			slog.String("host", r.Host),
			slog.String("request", r.RequestURI),
			slog.String("query", r.URL.RawQuery),
			slog.String("method", r.Method),
			slog.String("status", fmt.Sprintf("%d", rec.status)),
			slog.String("userAgent", r.UserAgent()),
			slog.String("referer", r.Referer()),
		}
		if _, ok := h.(*httputil.ReverseProxy); ok {
			h.(*httputil.ReverseProxy).Director(r)
			attrs = append(attrs, slog.String("target", fmt.Sprintf("proxy:%s", r.URL.String())))
		}
		log.LogAttrs(r.Context(), slog.LevelInfo, "access log", attrs...)
	})
}

func enforceAccessTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := getAccessToken(r)
		if accessToken == "" {
			slog.Error("No Authorization header, returning 401")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !introspectAndAddHeaders(r, accessToken) {
			slog.Error("Introspection failed, returning 401")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getAccessToken(req *http.Request) string {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = req.Header.Get("authorization")
	}
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

func introspectAndAddHeaders(req *http.Request, token string) bool {

	introspectionURL := getEnv("INTROSPECTION_URL", "http://localhost:3000/token/introspection")
	clientID := getEnv("CLIENT_ID", "client")
	clientSecret := getEnv("CLIENT_SECRET", "12345678")

	data := url.Values{}
	data.Set("token", token)

	client := &http.Client{}
	introspectionReq, err := http.NewRequest("POST", introspectionURL, strings.NewReader(data.Encode()))
	if err != nil {
		slog.Error("Failed to create introspection request", slog.String("error", err.Error()))
		return false
	}

	introspectionReq.SetBasicAuth(clientID, clientSecret)
	introspectionReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(introspectionReq)
	if err != nil {
		slog.Error("Failed to introspect token", slog.String("error", err.Error()))
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Token introspection returned non-200 status", slog.String("status", resp.Status))
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read introspection response body", slog.String("error", err.Error()))
		return false
	}

	var introspectionResponse map[string]interface{}
	if err := json.Unmarshal(body, &introspectionResponse); err != nil {
		slog.Error("Failed to unmarshal introspection response", slog.String("error", err.Error()))
		return false
	}

	if active, ok := introspectionResponse["active"].(bool); !ok || !active {
		slog.Error("Token is not active")
		return false
	}

	// Check x5t#S256 against client certificate's SHA-256 thumbprint
	if cnf, ok := introspectionResponse["cnf"].(map[string]interface{}); ok {
		if x5tS256, ok := cnf["x5t#S256"].(string); ok {
			if !verifyCertificateThumbprint(req, x5tS256) {
				slog.Error("Client certificate thumbprint verification failed")
				return false
			}
		}
	}

	// Base64 encode the introspection response and set as header
	introspectionResponseBase64 := base64.StdEncoding.EncodeToString(body)
	req.Header.Set("X-Introspection-Response", introspectionResponseBase64)
	//To be compliant with the lambda - to be removed
	req.Header.Set("access_token", string(body))

	// Check for 'sub' property and fetch user info if present
	if _, ok := introspectionResponse["sub"].(string); ok {
		fetchAndAddUserInfo(req, token)
	}

	return true
}

func fetchAndAddUserInfo(req *http.Request, token string) {
	userInfoURL := getEnv("USER_INFO_URL", "http://auth/me")

	client := &http.Client{}
	userInfoReq, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		slog.Error("Failed to create user info request", slog.String("error", err.Error()))
		return
	}

	userInfoReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(userInfoReq)
	if err != nil {
		slog.Error("Failed to fetch user info", slog.String("error", err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("User info request returned non-200 status", slog.String("status", resp.Status))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read user info response body", slog.String("error", err.Error()))
		return
	}

	// Base64 encode the user info response and set as header
	userInfoResponseBase64 := base64.StdEncoding.EncodeToString(body)
	req.Header.Set("X-User-Info-Response", userInfoResponseBase64)
}

func verifyCertificateThumbprint(req *http.Request, x5tS256 string) bool {
	if len(req.TLS.PeerCertificates) == 0 {
		return false
	}

	clientCert := req.TLS.PeerCertificates[0]
	hash := sha256.Sum256(clientCert.Raw)
	certThumbprint := base64.RawURLEncoding.EncodeToString(hash[:])

	return certThumbprint == x5tS256
}
