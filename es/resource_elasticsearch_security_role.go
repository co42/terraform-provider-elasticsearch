// Manage the role in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role.html
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

// Role object returned by API
type Role map[string]*RoleSpec

// RoleSpec is role object
type RoleSpec struct {
	Cluster      []string                    `json:"cluster"`
	Applications []RoleApplicationPrivileges `json:"applications,omitempty"`
	Indices      []RoleIndicesPermissions    `json:"indices,omitempty"`
	RunAs        []string                    `json:"run_as,omitempty"`
	Global       interface{}                 `json:"global,omitempty"`
	Metadata     interface{}                 `json:"metadata,omitempty"`
}

// RoleApplicationPrivileges is the application object
type RoleApplicationPrivileges struct {
	Application string   `json:"application"`
	Privileges  []string `json:"privileges,omitempty"`
	Resources   []string `json:"resources,omitempty"`
}

// RoleIndicesPermissions is the indice object
type RoleIndicesPermissions struct {
	Names         []string    `json:"names"`
	Privileges    []string    `json:"privileges"`
	FieldSecurity interface{} `json:"field_security,omitempty"`
	Query         interface{} `json:"query,omitempty"`
}

// resourceElasticsearchSecurityRole handle the role API call
func resourceElasticsearchSecurityRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSecurityRoleCreate,
		Read:   resourceElasticsearchSecurityRoleRead,
		Update: resourceElasticsearchSecurityRoleUpdate,
		Delete: resourceElasticsearchSecurityRoleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"run_as": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"global": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "{}",
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"indices": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"names": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"privileges": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"query": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "{}",
							DiffSuppressFunc: suppressEquivalentJSON,
						},
						"field_security": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "{}",
							DiffSuppressFunc: suppressEquivalentJSON,
						},
					},
				},
			},
			"applications": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application": {
							Type:     schema.TypeString,
							Required: true,
						},
						"privileges": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"resources": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

// resourceElasticsearchSecurityRoleCreate create new role in Elasticsearch
func resourceElasticsearchSecurityRoleCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	err := createRole(d, meta)
	if err != nil {
		return err
	}
	d.SetId(name)

	log.Infof("Created role %s successfully", name)

	return resourceElasticsearchSecurityRoleRead(d, meta)
}

// resourceElasticsearchSecurityRoleRead read existing role in Elasticsearch
func resourceElasticsearchSecurityRoleRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	log.Debugf("Role id:  %s", id)

	client := meta.(*elastic.Client)
	res, err := client.API.Security.GetRole(
		client.API.Security.GetRole.WithContext(context.Background()),
		client.API.Security.GetRole.WithPretty(),
		client.API.Security.GetRole.WithName(id),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Role %s not found - removing from state", id)
			log.Warnf("Role %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when get role %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.Debugf("Get role %s successfully:\n%s", id, string(b))
	role := make(Role)
	err = json.Unmarshal(b, &role)
	if err != nil {
		return err
	}

	log.Debugf("Role %+v", role)

	d.Set("name", id)
	d.Set("indices", role[id].Indices)
	d.Set("cluster", role[id].Cluster)
	d.Set("applications", role[id].Applications)
	d.Set("global", role[id].Global)
	d.Set("run_as", role[id].RunAs)
	d.Set("metadata", role[id].Metadata)

	log.Infof("Read role %s successfully", id)

	return nil
}

// resourceElasticsearchSecurityRoleUpdate update existing role in Elasticsearch
func resourceElasticsearchSecurityRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createRole(d, meta)
	if err != nil {
		return err
	}

	log.Infof("Updated role %s successfully", d.Id())

	return resourceElasticsearchSecurityRoleRead(d, meta)
}

// resourceElasticsearchSecurityRoleDelete delete existing role in Elasticsearch
func resourceElasticsearchSecurityRoleDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	log.Debugf("Role id: %s", id)

	client := meta.(*elastic.Client)
	res, err := client.API.Security.DeleteRole(
		id,
		client.API.Security.DeleteRole.WithContext(context.Background()),
		client.API.Security.DeleteRole.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Role %s not found - removing from state", id)
			log.Warnf("Role %s not found - removing from state", id)
			d.SetId("")
			return nil

		}
		return errors.Errorf("Error when delete role %s: %s", id, res.String())
	}

	d.SetId("")

	log.Infof("Deleted role %s successfully", id)
	return nil

}

// Print Role object as Json string
func (r *RoleSpec) String() string {
	json, _ := json.Marshal(r)
	return string(json)
}

// createRole create or update role in Elasticsearch
func createRole(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	indices := buildRolesIndicesPermissions(d.Get("indices").(*schema.Set).List())
	applications := buildRolesApplicationPrivileges(d.Get("applications").(*schema.Set).List())
	cluster := convertArrayInterfaceToArrayString(d.Get("cluster").(*schema.Set).List())
	global := optionalInterfaceJSON(d.Get("global").(string))
	runAs := convertArrayInterfaceToArrayString(d.Get("run_as").(*schema.Set).List())
	metadata := optionalInterfaceJSON(d.Get("metadata").(string))

	role := &RoleSpec{
		Cluster:      cluster,
		Applications: applications,
		Indices:      indices,
		RunAs:        runAs,
		Global:       global,
		Metadata:     metadata,
	}
	log.Debug("Name: ", name)
	log.Debug("Role: ", role)

	data, err := json.Marshal(role)
	if err != nil {
		return err
	}

	client := meta.(*elastic.Client)
	res, err := client.API.Security.PutRole(
		name,
		bytes.NewReader(data),
		client.API.Security.PutRole.WithContext(context.Background()),
		client.API.Security.PutRole.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add role %s: %s", name, res.String())
	}

	return nil
}

// buildRolesIndicesPermissions convert list to list of RoleIndicesPermissions objects
func buildRolesIndicesPermissions(raws []interface{}) []RoleIndicesPermissions {
	rolesIndicesPermissions := make([]RoleIndicesPermissions, len(raws))

	for i, raw := range raws {
		m := raw.(map[string]interface{})
		roleIndicesPermisions := RoleIndicesPermissions{
			Names:         convertArrayInterfaceToArrayString(m["names"].(*schema.Set).List()),
			Privileges:    convertArrayInterfaceToArrayString(m["privileges"].(*schema.Set).List()),
			Query:         optionalInterfaceJSON(m["query"].(string)),
			FieldSecurity: optionalInterfaceJSON(m["field_security"].(string)),
		}

		rolesIndicesPermissions[i] = roleIndicesPermisions

	}

	return rolesIndicesPermissions
}

// buildRolesApplicationPrivileges convert list to list of RoleApplicationPrivileges objects
func buildRolesApplicationPrivileges(raws []interface{}) []RoleApplicationPrivileges {
	rolesApplicationPrivileges := make([]RoleApplicationPrivileges, len(raws))

	for i, raw := range raws {
		m := raw.(map[string]interface{})
		roleApplicationPrivileges := RoleApplicationPrivileges{
			Application: m["application"].(string),
			Privileges:  convertArrayInterfaceToArrayString(m["privileges"].(*schema.Set).List()),
			Resources:   convertArrayInterfaceToArrayString(m["resources"].(*schema.Set).List()),
		}

		rolesApplicationPrivileges[i] = roleApplicationPrivileges

	}

	return rolesApplicationPrivileges
}
