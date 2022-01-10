package es

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Provider permiit to init the terraform provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"urls": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_URLS", nil),
				Description: "Elasticsearch URLs",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_USERNAME", nil),
				Description: "Username to use to connect to elasticsearch using basic auth",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_PASSWORD", nil),
				Description: "Password to use to connect to elasticsearch using basic auth",
			},
			"cacert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A Custom CA certificate",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable SSL verification of API calls",
			},
			"retry": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     6,
				Description: "Nummber time it retry connexion before failed",
			},
			"wait_before_retry": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "Wait time in second before retry connexion",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"elasticsearch_index_lifecycle_policy":    resourceElasticsearchIndexLifecyclePolicy(),
			"elasticsearch_index_template":            resourceElasticsearchIndexTemplate(),
			"elasticsearch_role":                      resourceElasticsearchSecurityRole(),
			"elasticsearch_role_mapping":              resourceElasticsearchSecurityRoleMapping(),
			"elasticsearch_user":                      resourceElasticsearchSecurityUser(),
			"elasticsearch_license":                   resourceElasticsearchLicense(),
			"elasticsearch_snapshot_repository":       resourceElasticsearchSnapshotRepository(),
			"elasticsearch_snapshot_lifecycle_policy": resourceElasticsearchSnapshotLifecyclePolicy(),
			"elasticsearch_watcher":                   resourceElasticsearchWatcher(),
		},

		ConfigureFunc: providerConfigure,
	}
}

// providerConfigure permit to initialize the rest client to access on Elasticsearch API
func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	var (
		data map[string]interface{}
	)

	URLs := strings.Split(d.Get("urls").(string), ",")
	insecure := d.Get("insecure").(bool)
	cacertFile := d.Get("cacert_file").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	retry := d.Get("retry").(int)
	waitBeforeRetry := d.Get("wait_before_retry").(int)
	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{},
	}
	// Checks is valid URLs
	for _, rawURL := range URLs {
		_, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
	}

	// Intialise connexion
	cfg := elastic.Config{
		Addresses: URLs,
	}
	if username != "" && password != "" {
		cfg.Username = username
		cfg.Password = password
	}
	if insecure == true {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}
	// If a cacertFile has been specified, use that for cert validation
	if cacertFile != "" {
		caCert, _, _ := read(cacertFile)

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		transport.TLSClientConfig.RootCAs = caCertPool
	}
	cfg.Transport = transport
	client, err := elastic.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Test connexion and check elastic version to use the right Version
	nbFailed := 0
	isOnline := false
	var res *esapi.Response
	for isOnline == false {
		res, err = client.API.Info(
			client.API.Info.WithContext(context.Background()),
		)
		if err == nil && res.IsError() == false {
			isOnline = true
		} else {
			if nbFailed == retry {
				return nil, err
			}
			nbFailed++
			time.Sleep(time.Duration(waitBeforeRetry) * time.Second)
		}
	}

	defer res.Body.Close()
	if res.IsError() {
		return nil, errors.Errorf("Error when get info about Elasticsearch client: %s", res.String())
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	version := data["version"].(map[string]interface{})["number"].(string)
	log.Debugf("Server: %s", version)

	if version < "7.0.0" || version >= "8.0.0" {
		return nil, errors.Errorf("ElasticSearch version is not 7.x (%s), you need to use the right version of elasticsearch provider", version)
	}

	return client, nil
}

// If the argument is a path, Read loads it and returns the contents,
// otherwise the argument is assumed to be the desired contents and is simply
// returned.
//
// The boolean second return value can be called `wasPath` - it indicates if a
// path was detected and a file loaded.
func read(poc string) (string, bool, error) {
	if len(poc) == 0 {
		return poc, false, nil
	}

	path := poc
	if path[0] == '~' {
		var err error
		path, err = homedir.Expand(path)
		if err != nil {
			return path, true, err
		}
	}

	if _, err := os.Stat(path); err == nil {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return string(contents), true, err
		}
		return string(contents), true, nil
	}

	return poc, false, nil
}
