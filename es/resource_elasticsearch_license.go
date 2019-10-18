// Manage license in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/update-license.html
// Supported version:
//  - v6
//  - v7

package es

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// License object
type License map[string]*LicenseSpec

// LicenseSpec is license object
type LicenseSpec struct {
	UID                string  `json:"uid"`
	Type               string  `json:"type"`
	IssueDateInMillis  float64 `json:"issue_date_in_millis"`
	ExpiryDateInMillis float64 `json:"expiry_date_in_millis"`
	MaxNodes           float64 `json:"max_nodes"`
	IssuedTo           string  `json:"issued_to"`
	Issuer             string  `json:"issuer"`
	Signature          string  `json:"signature,omitempty"`
	StartDateInMillis  float64 `json:"start_date_in_millis"`
}

// resourceElasticsearchLicense handle the license API call
func resourceElasticsearchLicense() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchLicenseCreate,
		Read:   resourceElasticsearchLicenseRead,
		Update: resourceElasticsearchLicenseUpdate,
		Delete: resourceElasticsearchLicenseDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"license": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressLicense,
			},
			"use_basic_license": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"basic_license": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

// resourceElasticsearchLicenseCreate create license or enable basic license
func resourceElasticsearchLicenseCreate(d *schema.ResourceData, meta interface{}) error {
	err := createLicense(d, meta)
	if err != nil {
		return err
	}
	d.SetId("license")
	return resourceElasticsearchLicenseRead(d, meta)
}

// resourceElasticsearchLicense update license
func resourceElasticsearchLicenseUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createLicense(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchLicenseRead(d, meta)
}

// resourceElasticsearchLicenseRead read license
func resourceElasticsearchLicenseRead(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*elastic.Client)
	res, err := client.API.License.Get(
		client.API.License.Get.WithContext(context.Background()),
		client.API.License.Get.WithPretty(),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] License not found - removing from state")
			log.Warnf("License not found - removing from state")
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when get license: %s", res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.Debugf("Get license successfully:\n%s", string(b))

	license := make(License)
	err = json.Unmarshal(b, &license)
	if err != nil {
		return err
	}

	licenseSpec := license["license"]

	log.Debugf("License object: %s", licenseSpec.String())

	if licenseSpec.Type == "basic" {
		d.Set("basic_license", licenseSpec.String())
		d.Set("use_basic_license", true)
	} else {
		d.Set("license", licenseSpec.String())
		d.Set("use_basic_license", false)
	}

	return nil
}

// resourceElasticsearchLicenseDelete delete license
func resourceElasticsearchLicenseDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*elastic.Client)
	res, err := client.API.License.Delete(
		client.API.License.Delete.WithContext(context.Background()),
		client.API.License.Delete.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] License not found - removing from state")
			log.Warnf("License not found - removing from state")
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when delete license: %s", res.String())

	}

	d.SetId("")
	return nil
}

// createLicense add or update license
func createLicense(d *schema.ResourceData, meta interface{}) error {
	license := d.Get("license").(string)
	useBasicLicense := d.Get("use_basic_license").(bool)

	client := meta.(*elastic.Client)
	var err error
	var res *esapi.Response
	// Use enterprise lisence
	if useBasicLicense == false {
		log.Debug("Use enterprise license")
		res, err = client.API.License.Post(
			client.API.License.Post.WithContext(context.Background()),
			client.API.License.Post.WithPretty(),
			client.API.License.Post.WithAcknowledge(true),
			client.API.License.Post.WithBody(strings.NewReader(license)),
		)
	} else {
		// Use basic lisence if needed (basic license not yet enabled)
		log.Debug("Use basic license")
		res, err = client.API.License.GetBasicStatus(
			client.API.License.GetBasicStatus.WithContext(context.Background()),
			client.API.License.GetBasicStatus.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			return errors.Errorf("Error when check if basic license can be enabled: %s", res.String())
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		log.Debugf("Result when get basic license status: %s", string(b))

		data := make(map[string]interface{})
		err = json.Unmarshal(b, &data)
		if err != nil {
			return err
		}

		if data["eligible_to_start_basic"].(bool) == false {
			log.Infof("Basic license is already enabled")
			return nil
		}
		res, err = client.API.License.PostStartBasic(
			client.API.License.PostStartBasic.WithContext(context.Background()),
			client.API.License.PostStartBasic.WithPretty(),
			client.API.License.PostStartBasic.WithAcknowledge(true),
		)

	}

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add license: %s", res.String())
	}

	return nil
}

// Print License object as Json string
func (r *LicenseSpec) String() string {
	json, _ := json.Marshal(r)
	return string(json)
}
