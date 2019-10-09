// Manage snapshot lifecycle policy in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/slm-api-put.html
// Supported version:
//  - v7
package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	elastic7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Snapshot repository object
type SnapshotLifecyclePolicy map[string]SnapshotLifecyclePolicyGet
type SnapshotLifecyclePolicySpec struct {
	Schedule   string      `json:"schedule"`
	Name       string      `json:"name"`
	Repository string      `json:"repository"`
	Configs    interface{} `json:"config,omitempty"`
}
type SnapshotLifecyclePolicyGet struct {
	Policy *SnapshotLifecyclePolicySpec `json:"policy"`
}

func resourceElasticsearchSnapshotLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSnapshotLifecyclePolicyCreate,
		Read:   resourceElasticsearchSnapshotLifecyclePolicyRead,
		Update: resourceElasticsearchSnapshotLifecyclePolicyUpdate,
		Delete: resourceElasticsearchSnapshotLifecyclePolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"snapshot_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"configs": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceElasticsearchSnapshotLifecyclePolicyCreate(d *schema.ResourceData, meta interface{}) error {

	name := d.Get("name").(string)

	err := createSnapshotLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	d.SetId(name)
	return resourceElasticsearchSnapshotLifecyclePolicyRead(d, meta)
}

func resourceElasticsearchSnapshotLifecyclePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createSnapshotLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchSnapshotLifecyclePolicyRead(d, meta)
}

func resourceElasticsearchSnapshotLifecyclePolicyRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	var b []byte

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.SlmGetLifecycle(
			client.API.SlmGetLifecycle.WithContext(context.Background()),
			client.API.SlmGetLifecycle.WithPretty(),
			client.API.SlmGetLifecycle.WithPolicyID(id),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Snapshot lifecycle policy %s not found - removing from state", id)
				log.Warnf("Snapshot lifecycle policy %s not found - removing from state", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when get snapshot lifecycle policy %s: %s", id, res.String())
			}
		}
		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
	default:
		return errors.New("Snapshot lifecycle policy is only supported by the elastic library >= v7!")
	}

	log.Debugf("Get snapshot lifecycle policy successfully:\n%s", string(b))

	snapshotLifecyclePolicy := make(SnapshotLifecyclePolicy)
	err := json.Unmarshal(b, &snapshotLifecyclePolicy)
	if err != nil {
		return err
	}

	log.Debugf("SnapshotLifecyclePolicy object %+v", snapshotLifecyclePolicy)

	// Manage bug https://github.com/elastic/elasticsearch/issues/47664
	if len(snapshotLifecyclePolicy) == 0 {
		fmt.Printf("[WARN] Snapshot lifecycle policy %s not found - removing from state", id)
		log.Warnf("Snapshot lifecycle policy %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	d.Set("name", id)
	d.Set("snapshot_name", snapshotLifecyclePolicy[id].Policy.Name)
	d.Set("schedule", snapshotLifecyclePolicy[id].Policy.Schedule)
	d.Set("repository", snapshotLifecyclePolicy[id].Policy.Repository)
	d.Set("configs", snapshotLifecyclePolicy[id].Policy.Configs)

	return nil
}

func resourceElasticsearchSnapshotLifecyclePolicyDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.SlmDeleteLifecycle(
			id,
			client.API.SlmDeleteLifecycle.WithContext(context.Background()),
			client.API.SlmDeleteLifecycle.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Snapshot lifecycle policy %s not found - removing from state", id)
				log.Warnf("Snapshot lifecycle policy %s not found - removing from state", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when delete snapshot lifecycle policy %s: %s", id, res.String())
			}
		}
	default:
		return errors.New("Snapshot lifecycle policy is only supported by the elastic library >= v7!")
	}

	d.SetId("")
	return nil
}

// createSnapshotLifecyclePolicy permit to create or update snapshot lifecycle policy
func createSnapshotLifecyclePolicy(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	snapshotName := d.Get("snapshot_name").(string)
	schedule := d.Get("schedule").(string)
	repository := d.Get("repository").(string)
	configs := optionalInterfaceJson(d.Get("configs").(string))

	snapshotLifecyclePolicy := &SnapshotLifecyclePolicySpec{
		Name:       snapshotName,
		Schedule:   schedule,
		Repository: repository,
		Configs:    configs,
	}

	b, err := json.Marshal(snapshotLifecyclePolicy)
	if err != nil {
		return err
	}

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)

		res, err := client.API.SlmPutLifecycle(
			name,
			client.API.SlmPutLifecycle.WithBody(bytes.NewReader(b)),
			client.API.SlmPutLifecycle.WithContext(context.Background()),
			client.API.SlmPutLifecycle.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add snapshot lifecycle policy %s: %s", name, res.String())
		}
	default:
		return errors.New("Snapshot lifecyle policy is only supported by the elastic library >= v7!")
	}

	return nil
}

// Print snapshot lifecycle policy object as Json string
func (r *SnapshotLifecyclePolicySpec) String() string {
	json, _ := json.Marshal(r)
	return string(json)
}
