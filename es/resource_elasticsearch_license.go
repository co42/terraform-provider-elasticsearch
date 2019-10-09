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

	elastic6 "github.com/elastic/go-elasticsearch/v6"
	esapi6 "github.com/elastic/go-elasticsearch/v6/esapi"
	elastic7 "github.com/elastic/go-elasticsearch/v7"
	esapi7 "github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

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

func resourceElasticsearchLicenseCreate(d *schema.ResourceData, meta interface{}) error {
	err := createLicense(d, meta)
	if err != nil {
		return err
	}
	d.SetId("license")
	return resourceElasticsearchLicenseRead(d, meta)
}

func resourceElasticsearchLicenseUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createLicense(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchLicenseRead(d, meta)
}

func resourceElasticsearchLicenseRead(d *schema.ResourceData, meta interface{}) error {

	var b []byte

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	// V6
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.XPack.LicenseGet(
			client.API.XPack.LicenseGet.WithContext(context.Background()),
			client.API.XPack.LicenseGet.WithPretty(),
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
			} else {
				return errors.Errorf("Error when get license: %s", res.String())
			}

		}
		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
	// V7
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
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
			} else {
				return errors.Errorf("Error when get license: %s", res.String())
			}

		}
		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
	default:
		return errors.New("License is only supported by the elastic library >= v6!")
	}

	log.Debugf("Get license successfully:\n%s", string(b))

	license := make(map[string]interface{})
	err := json.Unmarshal(b, &license)
	if err != nil {
		return err
	}

	licenseSpec := license["license"].(map[string]interface{})
	if licenseSpec["type"].(string) == "basic" {
		d.Set("basic_license", string(b))
		d.Set("use_basic_license", true)
	} else {
		d.Set("license", string(b))
		d.Set("use_basic_license", false)
	}

	return nil
}

func resourceElasticsearchLicenseDelete(d *schema.ResourceData, meta interface{}) error {

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	// V6
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.XPack.LicenseDelete(
			client.API.XPack.LicenseDelete.WithContext(context.Background()),
			client.API.XPack.LicenseDelete.WithPretty(),
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
			} else {
				return errors.Errorf("Error when delete license: %s", res.String())
			}
		}
	// V7
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
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
			} else {
				return errors.Errorf("Error when delete license: %s", res.String())
			}
		}
	default:
		return errors.New("License is only supported by the elastic library >= v6!")
	}

	d.SetId("")
	return nil
}

func createLicense(d *schema.ResourceData, meta interface{}) error {
	license := d.Get("license").(string)
	useBasicLicense := d.Get("use_basic_license").(bool)

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	// V6
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		var err error
		var res *esapi6.Response
		// Use enterprise lisence
		if useBasicLicense == false {
			log.Debug("Use enterprise license")
			res, err = client.API.XPack.LicensePost(
				client.API.XPack.LicensePost.WithContext(context.Background()),
				client.API.XPack.LicensePost.WithPretty(),
				client.API.XPack.LicensePost.WithAcknowledge(true),
				client.API.XPack.LicensePost.WithBody(strings.NewReader(license)),
			)
		} else {
			// Use basic lisence if needed (basic license not yet enabled)
			log.Debug("Use basic license")
			res, err = client.API.XPack.LicenseGetBasicStatus(
				client.API.XPack.LicenseGetBasicStatus.WithContext(context.Background()),
				client.API.XPack.LicenseGetBasicStatus.WithPretty(),
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
			} else {
				res, err = client.API.XPack.LicensePostStartBasic(
					client.API.XPack.LicensePostStartBasic.WithContext(context.Background()),
					client.API.XPack.LicensePostStartBasic.WithPretty(),
					client.API.XPack.LicensePostStartBasic.WithAcknowledge(true),
				)
			}
		}

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add license: %s", res.String())
		}

	// V7
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		var err error
		var res *esapi7.Response
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
			} else {
				res, err = client.API.License.PostStartBasic(
					client.API.License.PostStartBasic.WithContext(context.Background()),
					client.API.License.PostStartBasic.WithPretty(),
					client.API.License.PostStartBasic.WithAcknowledge(true),
				)
			}
		}

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add license: %s", res.String())
		}
	default:
		return errors.New("License is only supported by the elastic library >= v6!")
	}

	return nil
}
