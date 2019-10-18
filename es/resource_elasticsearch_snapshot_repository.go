// Manage snapshot repository in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-snapshots.html
// Supported version:
//  - v6
//  - v7

package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// SnapshotRepository object returned by API
type SnapshotRepository map[string]*SnapshotRepositorySpec

// SnapshotRepositorySpec is the repository object
type SnapshotRepositorySpec struct {
	Type     string            `json:"type"`
	Settings map[string]string `json:"settings"`
}

// resourceElasticsearchSnapshotRepository handle the snapshot repository API call
func resourceElasticsearchSnapshotRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSnapshotRepositoryCreate,
		Read:   resourceElasticsearchSnapshotRepositoryRead,
		Update: resourceElasticsearchSnapshotRepositoryUpdate,
		Delete: resourceElasticsearchSnapshotRepositoryDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"settings": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// resourceElasticsearchSnapshotRepositoryCreate create snapshot repository
func resourceElasticsearchSnapshotRepositoryCreate(d *schema.ResourceData, meta interface{}) error {

	name := d.Get("name").(string)

	err := createSnapshotRepository(d, meta)
	if err != nil {
		return err
	}
	d.SetId(name)
	return resourceElasticsearchSnapshotRepositoryRead(d, meta)
}

// resourceElasticsearchSnapshotRepositoryUpdate update the snapshot repository
func resourceElasticsearchSnapshotRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createSnapshotRepository(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchSnapshotRepositoryRead(d, meta)
}

// resourceElasticsearchSnapshotRepositoryRead read the sanpshot repository
func resourceElasticsearchSnapshotRepositoryRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Snapshot.GetRepository(
		client.API.Snapshot.GetRepository.WithContext(context.Background()),
		client.API.Snapshot.GetRepository.WithPretty(),
		client.API.Snapshot.GetRepository.WithRepository(id),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Snapshot repository %s not found - removing from state", id)
			log.Warnf("Snapshot repository %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when get snapshot repository %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.Debugf("Get Snapshot repository successfully:\n%s", string(b))

	snapshotRepository := make(SnapshotRepository)
	err = json.Unmarshal(b, &snapshotRepository)
	if err != nil {
		return err
	}

	d.Set("name", id)
	d.Set("type", snapshotRepository[id].Type)
	d.Set("settings", snapshotRepository[id].Settings)

	return nil
}

// resourceElasticsearchSnapshotRepositoryDelete delete the snapshot repository
func resourceElasticsearchSnapshotRepositoryDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Snapshot.DeleteRepository(
		[]string{id},
		client.API.Snapshot.DeleteRepository.WithContext(context.Background()),
		client.API.Snapshot.DeleteRepository.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Snapshot repository %s not found - removing from state", id)
			log.Warnf("Snapshot repository %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when delete snapshot repository %s: %s", id, res.String())

	}

	d.SetId("")
	return nil
}

// createSnapshotRepository create or update snapshot repository
func createSnapshotRepository(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	snapshotType := d.Get("type").(string)
	settings := convertMapInterfaceToMapString(d.Get("settings").(map[string]interface{}))

	snapshotRepository := &SnapshotRepositorySpec{
		Type:     snapshotType,
		Settings: settings,
	}

	b, err := json.Marshal(snapshotRepository)
	if err != nil {
		return err
	}

	client := meta.(*elastic.Client)

	res, err := client.API.Snapshot.CreateRepository(
		name,
		bytes.NewReader(b),
		client.API.Snapshot.CreateRepository.WithContext(context.Background()),
		client.API.Snapshot.CreateRepository.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add snapshot repository %s: %s", name, res.String())
	}

	return nil
}

// Print snapshot repository object as Json string
func (r *SnapshotRepositorySpec) String() string {
	json, _ := json.Marshal(r)
	return string(json)
}
