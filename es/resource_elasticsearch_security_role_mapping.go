package es

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	elastic7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func resourceElasticsearchSecurityRoleMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSecurityRoleMappingCreate,
		Read:   resourceElasticsearchSecurityRoleMappingRead,
		Update: resourceElasticsearchSecurityRoleMappingUpdate,
		Delete: resourceElasticsearchSecurityRoleMappingDelete,

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

func resourceElasticsearchSecurityRoleMappingRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	var b []byte

	log.Debugf("Role mapping id:  %s", id)

	switch meta.(type) {
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
			return errors.Errorf("Error when get role mapping %s: %s", id, res.String())
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

func resourceElasticsearchSecurityRoleMappingUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createRoleMapping(d, meta)
	if err != nil {
		return err
	}

	log.Infof("Updated role mapping %s successfully", d.Id())
	return resourceElasticsearchSecurityRoleMappingRead(d, meta)
}

func resourceElasticsearchSecurityRoleMappingDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	log.Debugf("Role mapping id: %s", id)

	switch meta.(type) {
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
			return errors.Errorf("Error when delete role mapping %s: %s", id, res.String())
		}

	default:
		return errors.New("Role mapping is only supported by the elastic library >= v6!")
	}

	d.SetId("")

	log.Infof("Deleted role mapping %s successfully", id)
	return nil

}

type RoleMapping map[string]RoleMappingSpec

type RoleMappingSpec struct {
	Roles    []string    `json:"roles"`
	Enabled  bool        `json:"enabled"`
	Rules    interface{} `json:"rules,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
}

func (r *RoleMappingSpec) String() string {
	json, _ := json.Marshal(r)
	return string(json)
}

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

	switch meta.(type) {
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
