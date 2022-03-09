// Manage component template in Elasticsearch
// API documentation:https://www.elastic.co/guide/en/elasticsearch/reference/master/indices-component-template.html
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
	oelastic "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchIndexComponentTemplate handle the index component template API call
func resourceElasticsearchIndexComponentTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexComponentTemplateCreate,
		Update: resourceElasticsearchIndexComponentTemplateUpdate,
		Read:   resourceElasticsearchIndexComponentTemplateRead,
		Delete: resourceElasticsearchIndexComponentTemplateDelete,

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

// resourceElasticsearchIndexComponentTemplateCreate create index component template
func resourceElasticsearchIndexComponentTemplateCreate(d *schema.ResourceData, meta interface{}) error {

	err := createIndexComponentTemplate(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIndexComponentTemplateRead(d, meta)
}

// resourceElasticsearchIndexComponentTemplateUpdate update index component template
func resourceElasticsearchIndexComponentTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createIndexComponentTemplate(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIndexComponentTemplateRead(d, meta)
}

// resourceElasticsearchIndexComponentTemplateRead read index component template
func resourceElasticsearchIndexComponentTemplateRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Cluster.GetComponentTemplate(
		client.API.Cluster.GetComponentTemplate.WithName(id),
		client.API.Cluster.GetComponentTemplate.WithContext(context.Background()),
		client.API.Cluster.GetComponentTemplate.WithPretty(),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Index component template %s not found - removing from state", id)
			log.Warnf("Index component template %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when get index component template %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	indexComponentTemplateResp := &oelastic.IndicesGetComponentTemplateResponse{}
	if err := json.Unmarshal(b, indexComponentTemplateResp); err != nil {
		return err
	}

	if len(indexComponentTemplateResp.ComponentTemplates) == 0 {
		fmt.Printf("[WARN] Index component template %s not found - removing from state", id)
		log.Warnf("Index component template %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	indexComponentTemplateJSON, err := json.Marshal(indexComponentTemplateResp.ComponentTemplates[0].ComponentTemplate)
	if err != nil {
		return err
	}

	log.Debugf("Get index component template %s successfully:%+v", id, string(indexComponentTemplateJSON))
	d.Set("name", d.Id())
	d.Set("template", string(indexComponentTemplateJSON))
	return nil
}

// resourceElasticsearchIndexComponentTemplateDelete delete index template
func resourceElasticsearchIndexComponentTemplateDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Cluster.DeleteComponentTemplate(
		id,
		client.API.Cluster.DeleteComponentTemplate.WithContext(context.Background()),
		client.API.Cluster.DeleteComponentTemplate.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Index component template %s not found - removing from state", id)
			log.Warnf("Index component template %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when delete index component template %s: %s", id, res.String())

	}

	d.SetId("")
	return nil
}

// createIndexComponentTemplate create or update index component template
func createIndexComponentTemplate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	template := d.Get("template").(string)

	client := meta.(*elastic.Client)
	res, err := client.API.Cluster.PutComponentTemplate(
		name,
		strings.NewReader(template),
		client.API.Cluster.PutComponentTemplate.WithContext(context.Background()),
		client.API.Cluster.PutComponentTemplate.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add index component template %s: %s", name, res.String())
	}

	return nil
}
