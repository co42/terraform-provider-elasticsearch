// Manage index template in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-templates.html
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

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchIndexTemplateLegacy handle the index template API call
func resourceElasticsearchIndexTemplateLegacy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexTemplateLegacyCreate,
		Update: resourceElasticsearchIndexTemplateLegacyUpdate,
		Read:   resourceElasticsearchIndexTemplateLegacyRead,
		Delete: resourceElasticsearchIndexTemplateLegacyDelete,

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

// resourceElasticsearchIndexTemplateLegacyCreate create index template
func resourceElasticsearchIndexTemplateLegacyCreate(d *schema.ResourceData, meta interface{}) error {

	err := createIndexTemplateLegacy(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIndexTemplateLegacyRead(d, meta)
}

// resourceElasticsearchIndexTemplateLegacyUpdate update index template
func resourceElasticsearchIndexTemplateLegacyUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createIndexTemplateLegacy(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIndexTemplateLegacyRead(d, meta)
}

// resourceElasticsearchIndexTemplateLegacyRead read index template
func resourceElasticsearchIndexTemplateLegacyRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	client := meta.(*elastic.Client)
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
		}
		return errors.Errorf("Error when get index template %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	indexTemplate := make(map[string]*olivere.IndicesGetTemplateResponse)
	if err := json.Unmarshal(b, &indexTemplate); err != nil {
		return err
	}

	indexTemplateJSON, err := json.Marshal(indexTemplate[id])
	if err != nil {
		return err
	}

	log.Debugf("Get index template %s successfully:%+v", id, string(indexTemplateJSON))
	d.Set("name", d.Id())
	d.Set("template", string(indexTemplateJSON))
	return nil

}

// resourceElasticsearchIndexTemplateLegacyDelete delete index template
func resourceElasticsearchIndexTemplateLegacyDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	client := meta.(*elastic.Client)
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
		}
		return errors.Errorf("Error when delete index template %s: %s", id, res.String())

	}

	d.SetId("")
	return nil
}

// createIndexTemplateLegacy create or update index template
func createIndexTemplateLegacy(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	template := d.Get("template").(string)

	client := meta.(*elastic.Client)
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

	return nil
}
