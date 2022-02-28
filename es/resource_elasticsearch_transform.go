// Manage transform in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html
// Supported version:
//  - v7

package es

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type TransformGetResponse struct {
	Transforms []Transform `json:"transforms,omitempty"`
}

type Transform struct {
	Id          string            `json:"id"`
	Version     string            `json:"version"`
	CreateTime  int64             `json:"create_time"`
	Source      TransformSource   `json:"source"`
	Dest        TransformDest     `json:"dest"`
	Frequency   string            `json:"frequency"`
	Sync        TransformSync     `json:"sync"`
	Pivot       TransformPivot    `json:"pivot"`
	Description string            `json:"description"`
	Settings    TransformSettings `json:"settings"`
}

type TransformSource struct {
	Index []string    `json:"index"`
	Query interface{} `json:"query,omitempty"`
}

type TransformDest struct {
	Index    string `json:"index"`
	Pipeline string `json:"pipeline"`
}

type TransformSync struct {
	Time TransformTime `json:"time"`
}

type TransformTime struct {
	Field string `json:"field"`
	Delay string `json:"delay"`
}

type TransformPivot struct {
	GroupBy      map[string]interface{} `json:"group_by"`
	Aggregations map[string]interface{} `json:"aggregations"`
}

type TransformSettings struct {
	MaxPageSearchSize int `json:"max_page_search_size"`
}

// resourceElasticsearchTransform handle the transform API call
func resourceElasticsearchTransform() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchTransformCreate,
		Read:   resourceElasticsearchTransformRead,
		Delete: resourceElasticsearchTransformDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"transform": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				DiffSuppressFunc: diffSuppressTransform,
			},
		},
	}
}

// resourceElasticsearchTransformCreate create transform
func resourceElasticsearchTransformCreate(d *schema.ResourceData, meta interface{}) error {

	err := createTransform(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchTransformRead(d, meta)
}

// resourceElasticsearchTransformUpdate update transform
func resourceElasticsearchTransformUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createTransform(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchTransformRead(d, meta)
}

// resourceElasticsearchTransformRead read transform
func resourceElasticsearchTransformRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.TransformGetTransform(
		client.API.TransformGetTransform.WithTransformID(id),
		client.API.TransformGetTransform.WithContext(context.Background()),
		client.API.TransformGetTransform.WithPretty(),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Transform %s not found - removing from state", id)
			log.Warnf("Transform %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when get transform %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	transform := TransformGetResponse{}
	if err := json.Unmarshal(b, &transform); err != nil {
		return err
	}

	if len(transform.Transforms) == 0 {
		fmt.Printf("[WARN] Transform %s not found - removing from state", id)
		log.Warnf("Transform %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	transformJSON, err := json.Marshal(transform.Transforms[0])
	if err != nil {
		return err
	}

	log.Debugf("Get transform %s successfully:%+v", id, transformJSON)
	d.Set("name", d.Id())
	d.Set("transform", string(transformJSON))
	return nil
}

// resourceElasticsearchTransformDelete delete transform
func resourceElasticsearchTransformDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.TransformDeleteTransform(
		id,
		client.API.TransformDeleteTransform.WithContext(context.Background()),
		client.API.TransformDeleteTransform.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Transform %s not found - removing from state", id)
			log.Warnf("Transform %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when delete transform %s: %s", id, res.String())

	}

	d.SetId("")
	return nil
}

// createTransform create or update transform
func createTransform(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	transform := d.Get("transform").(string)

	client := meta.(*elastic.Client)
	res, err := client.API.TransformPutTransform(
		strings.NewReader(transform),
		name,
		client.API.TransformPutTransform.WithContext(context.Background()),
		client.API.TransformPutTransform.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add transform %s: %s", name, res.String())
	}

	return nil
}
