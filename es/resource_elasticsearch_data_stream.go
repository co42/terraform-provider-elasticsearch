// Manage data stream in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html
// Supported version:
//  - v7

package es

import (
	"context"
	"encoding/json"
	"fmt"
	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type IndicesGetDataStreamResponse struct {
	DataStreams []interface{} `json:"data_streams,omitempty"`
}

// resourceElasticsearchDataStream handle the data stream API call
func resourceElasticsearchDataStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchDataStreamCreate,
		Read:   resourceElasticsearchDataStreamRead,
		Delete: resourceElasticsearchDataStreamDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

// resourceElasticsearchDataStreamCreate create data stream
func resourceElasticsearchDataStreamCreate(d *schema.ResourceData, meta interface{}) error {

	err := createDataStream(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchDataStreamRead(d, meta)
}

// resourceElasticsearchDataStreamRead read data stream
func resourceElasticsearchDataStreamRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Indices.GetDataStream(
		client.API.Indices.GetDataStream.WithName(id),
		client.API.Indices.GetDataStream.WithContext(context.Background()),
		client.API.Indices.GetDataStream.WithPretty(),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Data stream %s not found - removing from state", id)
			log.Warnf("Data stream %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when get data stream %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	dataStream := IndicesGetDataStreamResponse{}
	if err := json.Unmarshal(b, &dataStream); err != nil {
		return err
	}

	if len(dataStream.DataStreams) == 0 {
		fmt.Printf("[WARN] Data stream %s not found - removing from state", id)
		log.Warnf("Data stream %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	dataStreamJSON, err := json.Marshal(dataStream.DataStreams[0])
	if err != nil {
		return err
	}

	log.Debugf("Get data stream %s successfully:%+v", id, dataStreamJSON)
	d.Set("name", d.Id())
	return nil
}

// resourceElasticsearchDataStreamDelete delete data stream
func resourceElasticsearchDataStreamDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Indices.DeleteDataStream(
		[]string{id},
		client.API.Indices.DeleteDataStream.WithContext(context.Background()),
		client.API.Indices.DeleteDataStream.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Data stream %s not found - removing from state", id)
			log.Warnf("Data stream %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when delete data stream %s: %s", id, res.String())

	}

	d.SetId("")
	return nil
}

// createDataStream create a data stream
func createDataStream(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	client := meta.(*elastic.Client)
	res, err := client.API.Indices.CreateDataStream(
		name,
		client.API.Indices.CreateDataStream.WithContext(context.Background()),
		client.API.Indices.CreateDataStream.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add data stream %s: %s", name, res.String())
	}

	return nil
}
