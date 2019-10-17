// Manage index lifecylce policy in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-put-lifecycle.html
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

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchIndexLifecyclePolicy handle the index lifecycle policy API call
func resourceElasticsearchIndexLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexLifecyclePolicyCreate,
		Read:   resourceElasticsearchIndexLifecyclePolicyRead,
		Update: resourceElasticsearchIndexLifecyclePolicyUpdate,
		Delete: resourceElasticsearchIndexLifecyclePolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
		},
	}
}

// resourceElasticsearchIndexLifecyclePolicyCreate create new index lifecycle policy
func resourceElasticsearchIndexLifecyclePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	err := createIndexLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIndexLifecyclePolicyRead(d, meta)
}

// resourceElasticsearchIndexLifecyclePolicyUpdate update index lifecycle policy
func resourceElasticsearchIndexLifecyclePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createIndexLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIndexLifecyclePolicyRead(d, meta)
}

// resourceElasticsearchIndexLifecyclePolicyRead read index lifecycle policy
func resourceElasticsearchIndexLifecyclePolicyRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.ILM.GetLifecycle(
		client.API.ILM.GetLifecycle.WithContext(context.Background()),
		client.API.ILM.GetLifecycle.WithPretty(),
		client.API.ILM.GetLifecycle.WithPolicy(id),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Index lifecycle policy %s not found - removing from state", id)
			log.Warnf("Index lifecycle policy %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when get lifecycle policy %s: %s", id, res.String())
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.Debugf("Get life cycle policy %s successfully:\n%s", id, string(b))

	policyTemp := make(map[string]interface{})
	err = json.Unmarshal(b, &policyTemp)
	if err != nil {
		return err
	}
	policy := policyTemp[id].(map[string]interface{})["policy"]

	log.Debugf("Policy : %+v", policy)

	d.Set("name", id)
	d.Set("policy", policy)
	return nil
}

// resourceElasticsearchIndexLifecyclePolicyDelete delete index lifecycle policy
func resourceElasticsearchIndexLifecyclePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.ILM.DeleteLifecycle(
		id,
		client.API.ILM.DeleteLifecycle.WithContext(context.Background()),
		client.API.ILM.DeleteLifecycle.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Index lifecycle policy %s not found - removing from state", id)
			log.Warnf("Index lifecycle policy %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when delete lifecycle policy %s: %s", id, res.String())
	}

	d.SetId("")
	return nil
}

// createIndexLifecyclePolicy create or update index lifecycle policy
func createIndexLifecyclePolicy(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	policy := d.Get("policy").(string)

	client := meta.(*elastic.Client)
	res, err := client.API.ILM.PutLifecycle(
		name,
		client.API.ILM.PutLifecycle.WithContext(context.Background()),
		client.API.ILM.PutLifecycle.WithPretty(),
		client.API.ILM.PutLifecycle.WithBody(strings.NewReader(policy)),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add lifecycle policy %s: %s", name, res.String())
	}

	return nil
}
