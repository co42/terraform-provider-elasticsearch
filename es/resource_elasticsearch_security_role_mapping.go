// Manage the role mapping in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html
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

	elastic6 "github.com/elastic/go-elasticsearch/v6"
	elastic7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Role mapping object
type RoleMapping map[string]RoleMappingSpec
type RoleMappingSpec struct {
	Roles    []string    `json:"roles"`
	Enabled  bool        `json:"enabled"`
	Rules    interface{} `json:"rules,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
}

// Role mapping resource specification
func resourceElasticsearchSecurityRoleMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSecurityRoleMappingCreate,
		Read:   resourceElasticsearchSecurityRoleMappingRead,
		Update: resourceElasticsearchSecurityRoleMappingUpdate,
		Delete: resourceElasticsearchSecurityRoleMappingDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"rules": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentJson,
			},
			"roles": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "{}",
				DiffSuppressFunc: suppressEquivalentJson,
			},
		},
	}
}

// Create new role mapping in Elasticsearch
func resourceElasticsearchSecurityRoleMappingCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	err := createRoleMapping(d, meta)
	if err != nil {
		return err
	}
	d.SetId(name)
	log.Infof("Created role mapping %s successfully", name)

	return resourceElasticsearchSecurityRoleMappingRead(d, meta)
}

// Read existing role mapping in Elasticsearch
func resourceElasticsearchSecurityRoleMappingRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	var b []byte

	log.Debugf("Role mapping id:  %s", id)

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	// v6
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.XPack.SecurityGetRoleMapping(
			client.API.XPack.SecurityGetRoleMapping.WithContext(context.Background()),
			client.API.XPack.SecurityGetRoleMapping.WithPretty(),
			client.API.XPack.SecurityGetRoleMapping.WithName(id),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Role mapping %s not found. Removing from state\n", id)
				log.Warnf("Role mapping %s not found. Removing from state\n", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when get role mapping %s: %s", id, res.String())
			}
		}
		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

	// v7
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.Security.GetRoleMapping(
			client.API.Security.GetRoleMapping.WithContext(context.Background()),
			client.API.Security.GetRoleMapping.WithPretty(),
			client.API.Security.GetRoleMapping.WithName(id),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Role mapping %s not found. Removing from state\n", id)
				log.Warnf("Role mapping %s not found. Removing from state\n", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when get role mapping %s: %s", id, res.String())
			}
		}
		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
	default:
		return errors.New("Role mapping is only supported by the elastic library >= v6!")
	}

	log.Debugf("Get role mapping %s successfully:\n%s", id, string(b))
	roleMapping := make(RoleMapping)
	err := json.Unmarshal(b, &roleMapping)
	if err != nil {
		return err
	}

	log.Debugf("Role mapping %+v", roleMapping)

	d.Set("name", id)
	d.Set("enabled", roleMapping[id].Enabled)
	d.Set("roles", roleMapping[id].Roles)
	d.Set("rules", roleMapping[id].Rules)
	d.Set("metadata", roleMapping[id].Metadata)

	log.Infof("Read role mapping %s successfully", id)
	return nil
}

// Update existing role mapping in Elasticsearch
func resourceElasticsearchSecurityRoleMappingUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createRoleMapping(d, meta)
	if err != nil {
		return err
	}

	log.Infof("Updated role mapping %s successfully", d.Id())

	return resourceElasticsearchSecurityRoleMappingRead(d, meta)
}

// Delete existing role mapping in Elasticsearch
func resourceElasticsearchSecurityRoleMappingDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	log.Debugf("Role mapping id: %s", id)

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	// v6
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.XPack.SecurityDeleteRoleMapping(
			id,
			client.API.XPack.SecurityDeleteRoleMapping.WithContext(context.Background()),
			client.API.XPack.SecurityDeleteRoleMapping.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Role mapping %s not found - removing from state", id)
				log.Warnf("Role mapping %s not found - removing from state", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when delete role mapping %s: %s", id, res.String())
			}
		}

	// v7
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.Security.DeleteRoleMapping(
			id,
			client.API.Security.DeleteRoleMapping.WithContext(context.Background()),
			client.API.Security.DeleteRoleMapping.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Role mapping %s not found - removing from state", id)
				log.Warnf("Role mapping %s not found - removing from state", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when delete role mapping %s: %s", id, res.String())
			}
		}

	default:
		return errors.New("Role mapping is only supported by the elastic library >= v6!")
	}

	d.SetId("")

	log.Infof("Deleted role mapping %s successfully", id)
	return nil

}

// Create or update role mapping
func createRoleMapping(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	enabled := d.Get("enabled").(bool)
	roles := convertArrayInterfaceToArrayString(d.Get("roles").(*schema.Set).List())
	rules := optionalInterfaceJson(d.Get("rules").(string))
	metadata := optionalInterfaceJson(d.Get("metadata").(string))

	roleMapping := &RoleMappingSpec{
		Enabled:  enabled,
		Roles:    roles,
		Rules:    rules,
		Metadata: metadata,
	}
	log.Debug("Name: ", name)
	log.Debug("RoleMapping: ", roleMapping)

	data, err := json.Marshal(roleMapping)
	if err != nil {
		return err
	}

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	// v6
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.XPack.SecurityPutRoleMapping(
			name,
			bytes.NewReader(data),
			client.API.XPack.SecurityPutRoleMapping.WithContext(context.Background()),
			client.API.XPack.SecurityPutRoleMapping.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add role mapping %s: %s", name, res.String())
		}

	// v7
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.Security.PutRoleMapping(
			name,
			bytes.NewReader(data),
			client.API.Security.PutRoleMapping.WithContext(context.Background()),
			client.API.Security.PutRoleMapping.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add role mapping %s: %s", name, res.String())
		}
	default:
		return errors.New("Role mapping is only supported by the elastic library >= v6!")
	}

	return nil
}

// Print role mapping as Json string
func (r *RoleMappingSpec) String() string {
	json, _ := json.Marshal(r)
	return string(json)
}
