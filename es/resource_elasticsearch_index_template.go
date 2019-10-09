// Manage index template in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-templates.html
// Supported version:
//  - v6
//  - v7

package es

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	elastic6 "github.com/elastic/go-elasticsearch/v6"
	elastic7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func resourceElasticsearchIndexTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexTemplateCreate,
		Update: resourceElasticsearchIndexTemplateUpdate,
		Read:   resourceElasticsearchIndexTemplateRead,
		Delete: resourceElasticsearchIndexTemplateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"template": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressIndexTemplate,
			},
		},
	}
}

func resourceElasticsearchIndexTemplateCreate(d *schema.ResourceData, meta interface{}) error {

	err := createIndexTemplate(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIndexTemplateRead(d, meta)
}

func resourceElasticsearchIndexTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createIndexTemplate(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIndexTemplateRead(d, meta)
}

func resourceElasticsearchIndexTemplateRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var body string

	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.Indices.GetTemplate(
			client.API.Indices.GetTemplate.WithName(id),
			client.API.Indices.GetTemplate.WithContext(context.Background()),
			client.API.Indices.GetTemplate.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Index template %s not found - removing from state", id)
				log.Warnf("Index template %s not found - removing from state", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when get index template %s: %s", id, res.String())
			}
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		body = string(b)
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.Indices.GetTemplate(
			client.API.Indices.GetTemplate.WithName(id),
			client.API.Indices.GetTemplate.WithContext(context.Background()),
			client.API.Indices.GetTemplate.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Index template %s not found - removing from state", id)
				log.Warnf("Index template %s not found - removing from state", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when get index template %s: %s", id, res.String())
			}
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		body = string(b)
	default:
		return errors.New("Index template is only supported by the elastic library >= v6!")
	}

	log.Debugf("Get index template %s successfully:\n%s", id, body)
	d.Set("name", d.Id())
	d.Set("template", body)
	return nil
}

func resourceElasticsearchIndexTemplateDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.Indices.DeleteTemplate(
			id,
			client.API.Indices.DeleteTemplate.WithContext(context.Background()),
			client.API.Indices.DeleteTemplate.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Index template %s not found - removing from state", id)
				log.Warnf("Index template %s not found - removing from state", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when delete index template %s: %s", id, res.String())
			}

		}
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.Indices.DeleteTemplate(
			id,
			client.API.Indices.DeleteTemplate.WithContext(context.Background()),
			client.API.Indices.DeleteTemplate.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when delete index template %s: %s", id, res.String())
		}
	default:
		return errors.New("Index template is only supported by the elastic library >= v6!")
	}

	d.SetId("")
	return nil
}

func createIndexTemplate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	template := d.Get("template").(string)

	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.Indices.PutTemplate(
			name,
			strings.NewReader(template),
			client.API.Indices.PutTemplate.WithContext(context.Background()),
			client.API.Indices.PutTemplate.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add index template %s: %s", name, res.String())
		}
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.Indices.PutTemplate(
			name,
			strings.NewReader(template),
			client.API.Indices.PutTemplate.WithContext(context.Background()),
			client.API.Indices.PutTemplate.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add index template %s: %s", name, res.String())
		}
	default:
		return errors.New("Index template is only supported by the elastic library >= v6!")
	}

	return nil
}
