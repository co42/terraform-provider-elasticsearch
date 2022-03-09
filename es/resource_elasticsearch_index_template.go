// Manage index template in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/index-templates.html
// Supported version:
//  - v7

package es

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchIndexTemplate handle the index template API call
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
				DiffSuppressFunc: suppressEquivalentJSON,
			},
		},
	}
}

// resourceElasticsearchIndexTemplateCreate create index template
func resourceElasticsearchIndexTemplateCreate(d *schema.ResourceData, meta interface{}) error {

	err := createIndexTemplate(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIndexTemplateRead(d, meta)
}

// resourceElasticsearchIndexTemplateUpdate update index template
func resourceElasticsearchIndexTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createIndexTemplate(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIndexTemplateRead(d, meta)
}

// resourceElasticsearchIndexTemplateRead read index template
func resourceElasticsearchIndexTemplateRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Indices.GetIndexTemplate(
		client.API.Indices.GetIndexTemplate.WithName(id),
		client.API.Indices.GetIndexTemplate.WithContext(context.Background()),
		client.API.Indices.GetIndexTemplate.WithPretty(),
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
		}
		return errors.Errorf("Error when get index template %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	indexTemplate := &olivere.IndicesGetIndexTemplateResponse{}
	if err := json.Unmarshal(b, indexTemplate); err != nil {
		return err
	}

	if len(indexTemplate.IndexTemplates) == 0 {
		fmt.Printf("[WARN] Index template %s not found - removing from state", id)
		log.Warnf("Index template %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	indexTemplateJSON, err := json.Marshal(indexTemplate.IndexTemplates[0].IndexTemplate)
	if err != nil {
		return err
	}

	log.Debugf("Get index template %s successfully:%+v", id, string(indexTemplateJSON))
	d.Set("name", d.Id())
	d.Set("template", string(indexTemplateJSON))
	return nil
}

// resourceElasticsearchIndexTemplateDelete delete index template
func resourceElasticsearchIndexTemplateDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Indices.DeleteIndexTemplate(
		id,
		client.API.Indices.DeleteIndexTemplate.WithContext(context.Background()),
		client.API.Indices.DeleteIndexTemplate.WithPretty(),
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
		}
		return errors.Errorf("Error when delete index template %s: %s", id, res.String())

	}

	d.SetId("")
	return nil
}

// createIndexTemplate create or update index template
func createIndexTemplate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	template := d.Get("template").(string)

	client := meta.(*elastic.Client)
	res, err := client.API.Indices.PutIndexTemplate(
		name,
		strings.NewReader(template),
		client.API.Indices.PutIndexTemplate.WithContext(context.Background()),
		client.API.Indices.PutIndexTemplate.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add index template %s: %s", name, res.String())
	}

	return nil
}
