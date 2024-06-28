package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry/cf_exporter/collectors"
	"github.com/cloudfoundry/cf_exporter/fetcher"
	"github.com/cloudfoundry/cf_exporter/filters"
	"github.com/prometheus/client_golang/prometheus"
	versionCollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	bbsAPIUrl = kingpin.Flag(
		"bbs.api_url", "BBS API URL ($CF_EXPORTER_BBS_API_URL)",
	).Envar("CF_EXPORTER_BBS_API_URL").String()

	bbsTimeout = kingpin.Flag(
		"bbs.timeout", "BBS API Timeout ($CF_EXPORTER_BBS_TIMEOUT)",
	).Envar("CF_EXPORTER_BBS_TIMEOUT").Default("10").Int()

	bbsCAFile = kingpin.Flag(
		"bbs.ca_file", "BBS CA File ($CF_EXPORTER_BBS_CA_FILE)",
	).Envar("CF_EXPORTER_BBS_CA_FILE").Default("").String()

	bbsCertFile = kingpin.Flag(
		"bbs.cert_file", "BBS Cert File ($CF_EXPORTER_BBS_CERT_FILE)",
	).Envar("CF_EXPORTER_BBS_CERT_FILE").Default("").String()

	bbsKeyFile = kingpin.Flag(
		"bbs.key_file", "BBS Key File ($CF_EXPORTER_BBS_KEY_FILE)",
	).Envar("CF_EXPORTER_BBS_KEY_FILE").String()

	bbsSkipSSLValidation = kingpin.Flag(
		"bbs.skip_ssl_verify", "Disable SSL Verify for BBS ($CF_EXPORTER_BBS_SKIP_SSL_VERIFY)",
	).Envar("CF_EXPORTER_BBS_SKIP_SSL_VERIFY").Default("false").Bool()

	cfAPIUrl = kingpin.Flag(
		"cf.api_url", "Cloud Foundry API URL ($CF_EXPORTER_CF_API_URL)",
	).Envar("CF_EXPORTER_CF_API_URL").String()

	cfUsername = kingpin.Flag(
		"cf.username", "Cloud Foundry Username ($CF_EXPORTER_CF_USERNAME)",
	).Envar("CF_EXPORTER_CF_USERNAME").String()

	cfPassword = kingpin.Flag(
		"cf.password", "Cloud Foundry Password ($CF_EXPORTER_CF_PASSWORD)",
	).Envar("CF_EXPORTER_CF_PASSWORD").String()

	cfClientID = kingpin.Flag(
		"cf.client-id", "Cloud Foundry Client ID ($CF_EXPORTER_CF_CLIENT_ID)",
	).Envar("CF_EXPORTER_CF_CLIENT_ID").String()

	cfClientSecret = kingpin.Flag(
		"cf.client-secret", "Cloud Foundry Client Secret ($CF_EXPORTER_CF_CLIENT_SECRET)",
	).Envar("CF_EXPORTER_CF_CLIENT_SECRET").String()

	cfDeploymentName = kingpin.Flag(
		"cf.deployment-name", "Cloud Foundry Deployment Name to be reported as a metric label ($CF_EXPORTER_CF_DEPLOYMENT_NAME)",
	).Envar("CF_EXPORTER_CF_DEPLOYMENT_NAME").Required().String()

	filterCollectors = kingpin.Flag(
		"filter.collectors", "Comma separated collectors to filter (Applications,Buildpacks,Events,IsolationSegments,Organizations,Routes,SecurityGroups,ServiceBindings,ServiceInstances,ServicePlans,Services,Spaces,Stacks,Tasks,ActualLRPs). If not set, all collectors except Events and Tasks are enabled ($CF_EXPORTER_FILTER_COLLECTORS)",
	).Envar("CF_EXPORTER_FILTER_COLLECTORS").Default("").String()

	metricsNamespace = kingpin.Flag(
		"metrics.namespace", "Metrics Namespace ($CF_EXPORTER_METRICS_NAMESPACE)",
	).Envar("CF_EXPORTER_METRICS_NAMESPACE").Default("cf").String()

	metricsEnvironment = kingpin.Flag(
		"metrics.environment", "Environment label to be attached to metrics ($CF_EXPORTER_METRICS_ENVIRONMENT)",
	).Envar("CF_EXPORTER_METRICS_ENVIRONMENT").Required().String()

	skipSSLValidation = kingpin.Flag(
		"skip-ssl-verify", "Disable SSL Verify ($CF_EXPORTER_SKIP_SSL_VERIFY)",
	).Envar("CF_EXPORTER_SKIP_SSL_VERIFY").Default("false").Bool()

	listenAddress = kingpin.Flag(
		"web.listen-address", "Address to listen on for web interface and telemetry ($CF_EXPORTER_WEB_LISTEN_ADDRESS)",
	).Envar("CF_EXPORTER_WEB_LISTEN_ADDRESS").Default(":9193").String()

	metricsPath = kingpin.Flag(
		"web.telemetry-path", "Path under which to expose Prometheus metrics ($CF_EXPORTER_WEB_TELEMETRY_PATH)",
	).Envar("CF_EXPORTER_WEB_TELEMETRY_PATH").Default("/metrics").String()

	authUsername = kingpin.Flag(
		"web.auth.username", "Username for web interface basic auth ($CF_EXPORTER_WEB_AUTH_USERNAME)",
	).Envar("CF_EXPORTER_WEB_AUTH_USERNAME").String()

	authPassword = kingpin.Flag(
		"web.auth.password", "Password for web interface basic auth ($CF_EXPORTER_WEB_AUTH_PASSWORD)",
	).Envar("CF_EXPORTER_WEB_AUTH_PASSWORD").String()

	tlsCertFile = kingpin.Flag(
		"web.tls.cert_file", "Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate ($CF_EXPORTER_WEB_TLS_CERTFILE)",
	).Envar("CF_EXPORTER_WEB_TLS_CERTFILE").ExistingFile()

	tlsKeyFile = kingpin.Flag(
		"web.tls.key_file", "Path to a file that contains the TLS private key (PEM format) ($CF_EXPORTER_WEB_TLS_KEYFILE)",
	).Envar("CF_EXPORTER_WEB_TLS_KEYFILE").ExistingFile()

	logLevel = kingpin.Flag(
		"log.level", "Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]",
	).Default("error").String()

	logOutput = kingpin.Flag(
		"log.stream", "Set output stream for log. Valid outputs: [stderr, stdout]",
	).Default("stdout").String()

	logJSON = kingpin.Flag(
		"log.json", "Output logs with JSON format",
	).Default("false").Bool()

	workers = kingpin.Flag(
		"collector.workers", "Number of requests threads for collector",
	).Default("10").Int()
)

func init() {
	prometheus.MustRegister(versionCollector.NewCollector(*metricsNamespace))
}

type basicAuthHandler struct {
	handler  http.HandlerFunc
	username string
	password string
}

func (h *basicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || username != h.username || password != h.password {
		log.Errorf("Invalid HTTP auth from `%s`", r.RemoteAddr)
		w.Header().Set("WWW-Authenticate", "Basic realm=\"metrics\"")
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	h.handler(w, r)
}

func prometheusHandler() http.Handler {
	handler := promhttp.Handler()
	if *authUsername != "" && *authPassword != "" {
		handler = &basicAuthHandler{
			handler:  promhttp.Handler().ServeHTTP,
			username: *authUsername,
			password: *authPassword,
		}
	}
	return handler
}

func main() {
	kingpin.Version(version.Print("cf_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting cf_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	if *logJSON {
		log.SetFormatter(&log.JSONFormatter{})
	}
	if *logOutput == "stderr" {
		log.SetOutput(os.Stderr)
	}
	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Warnf("invalid log.level=%s, defaulting to error", *logLevel)
		lvl = log.ErrorLevel
	}
	log.SetLevel(lvl)

	cfConfig := &fetcher.CFConfig{
		URL:               *cfAPIUrl,
		Username:          *cfUsername,
		Password:          *cfPassword,
		ClientID:          *cfClientID,
		ClientSecret:      *cfClientSecret,
		SkipSSLValidation: *skipSSLValidation,
	}

	bbsConfig := &fetcher.BBSConfig{
		URL:            *bbsAPIUrl,
		Timeout:        *bbsTimeout,
		CAFile:         *bbsCAFile,
		CertFile:       *bbsCertFile,
		KeyFile:        *bbsKeyFile,
		SkipCertVerify: *bbsSkipSSLValidation,
	}

	active := []string{}
	if len(*filterCollectors) != 0 {
		active = strings.Split(*filterCollectors, ",")
	}
	filter, err := filters.NewFilter(active...)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	c, err := collectors.NewCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, *workers, cfConfig, bbsConfig, filter)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	prometheus.MustRegister(c)

	handler := prometheusHandler()
	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(`<html>
             <head><title>Cloud Foundry Exporter</title></head>
             <body>
             <h1>Cloud Foundry Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	server := &http.Server{
		Addr:              *listenAddress,
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 10,
	}

	if *tlsCertFile != "" && *tlsKeyFile != "" {
		log.Infoln("Listening TLS on", *listenAddress)
		err = server.ListenAndServeTLS(*tlsCertFile, *tlsKeyFile)
	} else {
		log.Infoln("Listening on", *listenAddress)
		err = server.ListenAndServe()
	}

	log.Fatal(err)
}
