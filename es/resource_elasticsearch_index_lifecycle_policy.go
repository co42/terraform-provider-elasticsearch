// Manage lifecycle policy in Elastricsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-put-lifecycle.html
package es

import (
	"context"
	"io/ioutil"
	"strings"

	elastic6 "github.com/elastic/go-elasticsearch/v6"
	elastic7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type LifeCycleSpec struct{
	Policy *LifeCyclePolicy `json:"policy"`
}
type LifeCyclePolicy struct {
	Phases map[string]LifeCyclePhase `json:"phases"`
}
type LifeCyclePhase struct {
	MinAge string `json:"min_age"`
	Actions map[string]map[string]interface{} `json:"actions"`
}

// Resource Lifecycle policy specification
func resourceElasticsearchIndexLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexLifecyclePolicyCreate,
		Read:   resourceElasticsearchIndexLifecyclePolicyRead,
		Update: resourceElasticsearchIndexLifecyclePolicyUpdate,
		Delete: resourceElasticsearchIndexLifecyclePolicyDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"phase": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"min_age": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"actions": {
						Type:     schema.TypeSet,
						Required: true,
						Elem: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"options": {
								Type:     schema.TypeMap,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceElasticsearchIndexLifecyclePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	err := createIndexLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIndexLifecyclePolicyRead(d, meta)
}

func resourceElasticsearchIndexLifecyclePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createIndexLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIndexLifecyclePolicyRead(d, meta)
}

func resourceElasticsearchIndexLifecyclePolicyRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var body string

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
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
			return errors.Errorf("Error when get lifecycle policy %s: %s", id, res.String())
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		body = string(b)
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
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
			return errors.Errorf("Error when get lifecycle policy %s: %s", id, res.String())
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		body = string(b)
	default:
		return errors.New("Index Lifecycle Management is only supported by the elastic library >= v6!")
	}

	log.Debugf("Get life cycle policy %s successfully:\n%s", id, body)
	d.Set("name", d.Id())
	d.Set("policy", body)
	return nil
}

func resourceElasticsearchIndexLifecyclePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.ILM.DeleteLifecycle(
			client.API.ILM.DeleteLifecycle.WithContext(context.Background()),
			client.API.ILM.DeleteLifecycle.WithPretty(),
			client.API.ILM.DeleteLifecycle.WithPolicy(id),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when delete lifecycle policy %s: %s", id, res.String())
		}
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.ILM.DeleteLifecycle(
			client.API.ILM.DeleteLifecycle.WithContext(context.Background()),
			client.API.ILM.DeleteLifecycle.WithPretty(),
			client.API.ILM.DeleteLifecycle.WithPolicy(id),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when delete lifecycle policy %s: %s", id, res.String())
		}
	default:
		return errors.New("Index Lifecycle Management is only supported by the elastic library >= v6!")
	}

	d.SetId("")
	return nil
}

func createIndexLifecyclePolicy(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	policy := d.Get("policy").(string)

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.ILM.PutLifecycle(
			client.API.ILM.PutLifecycle.WithContext(context.Background()),
			client.API.ILM.PutLifecycle.WithPretty(),
			client.API.ILM.PutLifecycle.WithPolicy(name),
			client.API.ILM.PutLifecycle.WithBody(strings.NewReader(policy)),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add lifecycle policy %s: %s", name, res.String())
		}
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.ILM.PutLifecycle(
			client.API.ILM.PutLifecycle.WithContext(context.Background()),
			client.API.ILM.PutLifecycle.WithPretty(),
			client.API.ILM.PutLifecycle.WithPolicy(name),
			client.API.ILM.PutLifecycle.WithBody(strings.NewReader(policy)),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add lifecycle policy %s: %s", name, res.String())
		}
	default:
		return errors.New("Index Lifecycle Management is only supported by the elastic library >= v6!")
	}

	return nil
}
