// Manage ingest pipeline in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html
// Supported version:
//  - v7

package es

import (
	"context"
	"encoding/json"
	"fmt"
	olivere "github.com/olivere/elastic/v7"
	"io/ioutil"
	"strings"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchIngestPipeline handle the ingest pipeline API call
func resourceElasticsearchIngestPipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIngestPipelineCreate,
		Update: resourceElasticsearchIngestPipelineUpdate,
		Read:   resourceElasticsearchIngestPipelineRead,
		Delete: resourceElasticsearchIngestPipelineDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"pipeline": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressIngestPipeline,
			},
		},
	}
}

// resourceElasticsearchIngestPipelineCreate create ingest pipeline
func resourceElasticsearchIngestPipelineCreate(d *schema.ResourceData, meta interface{}) error {

	err := createIngestPipeline(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIngestPipelineRead(d, meta)
}

// resourceElasticsearchIngestPipelineUpdate update ingest pipeline
func resourceElasticsearchIngestPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createIngestPipeline(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIngestPipelineRead(d, meta)
}

// resourceElasticsearchIngestPipelineRead read ingest pipeline
func resourceElasticsearchIngestPipelineRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Ingest.GetPipeline(
		client.API.Ingest.GetPipeline.WithPipelineID(id),
		client.API.Ingest.GetPipeline.WithContext(context.Background()),
		client.API.Ingest.GetPipeline.WithPretty(),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Ingest pipeline %s not found - removing from state", id)
			log.Warnf("Ingest pipeline %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when get ingest pipeline %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	ingestPipeline := olivere.IngestGetPipelineResponse{}
	if err := json.Unmarshal(b, &ingestPipeline); err != nil {
		return err
	}

	if _, ok := ingestPipeline[id]; !ok {
		fmt.Printf("[WARN] Ingest pipeline %s not found - removing from state", id)
		log.Warnf("Ingest pipeline %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	ingestPipelineJSON, err := json.Marshal(ingestPipeline[id])
	if err != nil {
		return err
	}

	log.Debugf("Get ingest pipeline %s successfully:%+v", id, ingestPipelineJSON)
	d.Set("name", d.Id())
	d.Set("pipeline", string(ingestPipelineJSON))
	return nil
}

// resourceElasticsearchIngestPipelineDelete delete ingest pipeline
func resourceElasticsearchIngestPipelineDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Ingest.DeletePipeline(
		id,
		client.API.Ingest.DeletePipeline.WithContext(context.Background()),
		client.API.Ingest.DeletePipeline.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Ingest pipeline %s not found - removing from state", id)
			log.Warnf("Ingest pipeline %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when delete ingest pipeline %s: %s", id, res.String())

	}

	d.SetId("")
	return nil
}

// createIngestPipeline create or update ingest pipeline
func createIngestPipeline(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	pipeline := d.Get("pipeline").(string)

	client := meta.(*elastic.Client)
	res, err := client.API.Ingest.PutPipeline(
		name,
		strings.NewReader(pipeline),
		client.API.Ingest.PutPipeline.WithContext(context.Background()),
		client.API.Ingest.PutPipeline.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add ingest pipeline %s: %s", name, res.String())
	}

	return nil
}
