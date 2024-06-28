package fetcher

import (
	"strings"
	"time"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/trace"
	"code.cloudfoundry.org/lager/v3"
)

const (
	clientSessionCacheSize int = -1
	maxIdleConnsPerHost    int = -1
)

type BBSClient struct {
	client bbs.Client
	config *BBSConfig
	logger lager.Logger
}

type BBSConfig struct {
	URL            string `yaml:"url"`
	Timeout        int    `yaml:"timeout"`
	CAFile         string `yaml:"ca_file"`
	CertFile       string `yaml:"cert_file"`
	KeyFile        string `yaml:"key_file"`
	SkipCertVerify bool   `yaml:"skip_cert_verify"`
}

func NewBBSClient(config *BBSConfig) (*BBSClient, error) {
	var err error
	bbsClient := BBSClient{
		config: config,
		logger: lager.NewLogger("bbs-client"),
	}
	bbsClientConfig := bbs.ClientConfig{
		URL:            config.URL,
		Retries:        1,
		RequestTimeout: time.Duration(config.Timeout) * time.Second,
	}
	if strings.HasPrefix(config.URL, "https://") {
		bbsClientConfig.IsTLS = true
		bbsClientConfig.InsecureSkipVerify = config.SkipCertVerify
		bbsClientConfig.CAFile = config.CAFile
		bbsClientConfig.CertFile = config.CertFile
		bbsClientConfig.KeyFile = config.KeyFile
		bbsClientConfig.ClientSessionCacheSize = clientSessionCacheSize
		bbsClientConfig.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
	bbsClient.client, err = bbs.NewClientWithConfig(bbsClientConfig)
	return &bbsClient, err
}

func (b *BBSClient) GetActualLRPs() ([]*models.ActualLRP, error) {
	traceID := trace.GenerateTraceID()
	actualLRPs, err := b.client.ActualLRPs(b.logger, traceID, models.ActualLRPFilter{})

	return actualLRPs, err
}
