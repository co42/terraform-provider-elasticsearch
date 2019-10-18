// Manage the user in elasticsearch
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

// User Json object returned by API
type User map[string]*UserSpec

// UserSpec is the user object
type UserSpec struct {
	Enabled      bool        `json:"enabled"`
	Email        string      `json:"email"`
	FullName     string      `json:"full_name"`
	Password     string      `json:"password,omitempty"`
	PasswordHash string      `json:"password_hash,omitempty"`
	Roles        []string    `json:"roles"`
	Metadata     interface{} `json:"metadata,omitempty"`
}

// resourceElasticsearchSecurityUser handle the user API call
func resourceElasticsearchSecurityUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSecurityUserCreate,
		Read:   resourceElasticsearchSecurityUserRead,
		Update: resourceElasticsearchSecurityUserUpdate,
		Delete: resourceElasticsearchSecurityUserDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password_hash": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "{}",
				DiffSuppressFunc: suppressEquivalentJSON,
			},
		},
	}
}

// resourceElasticsearchSecurityUserCreate create new user in Elasticsearch
func resourceElasticsearchSecurityUserCreate(d *schema.ResourceData, meta interface{}) error {
	username := d.Get("username").(string)

	err := createUser(d, meta, false)
	if err != nil {
		return err
	}
	d.SetId(username)

	log.Infof("Created user %s successfully", username)

	return resourceElasticsearchSecurityUserRead(d, meta)
}

// resourceElasticsearchSecurityUserRead read existing user in Elasticsearch
func resourceElasticsearchSecurityUserRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	log.Debugf("User id:  %s", id)

	client := meta.(*elastic.Client)
	res, err := client.API.Security.GetUser(
		client.API.Security.GetUser.WithContext(context.Background()),
		client.API.Security.GetUser.WithPretty(),
		client.API.Security.GetUser.WithUsername(id),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] User %s not found - removing from state", id)
			log.Warnf("User %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when get user %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.Debugf("Get user %s successfully:\n%s", id, string(b))
	user := make(User)
	err = json.Unmarshal(b, &user)
	if err != nil {
		return err
	}

	log.Debugf("User %+v", user)

	d.Set("username", id)
	d.Set("enabled", user[id].Enabled)
	d.Set("email", user[id].Email)
	d.Set("full_name", user[id].FullName)
	d.Set("roles", user[id].Roles)
	d.Set("metadata", user[id].Metadata)

	log.Infof("Read user %s successfully", id)

	return nil
}

// resourceElasticsearchSecurityUserUpdate update existing user in Elasticsearch
func resourceElasticsearchSecurityUserUpdate(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	// Use change password API if needed
	if d.HasChange("password") || d.HasChange("password_hash") {

		payload := make(map[string]string)
		if d.HasChange("password") {
			payload["password"] = d.Get("password").(string)
		} else {
			payload["password_hash"] = d.Get("password_hash").(string)
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		client := meta.(*elastic.Client)
		res, err := client.API.Security.ChangePassword(
			bytes.NewReader(data),
			client.API.Security.ChangePassword.WithUsername(id),
			client.API.Security.ChangePassword.WithContext(context.Background()),
			client.API.Security.ChangePassword.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when change password for user %s: %s", id, res.String())
		}

		log.Infof("Updated user password %s successfully", d.Id())

	}

	// Use user API for other fiedls
	if d.HasChange("enabled") || d.HasChange("email") || d.HasChange("full_name") || d.HasChange("roles") || d.HasChange("metadata") {
		err := createUser(d, meta, true)
		if err != nil {
			return err
		}

		log.Infof("Updated user %s successfully", d.Id())

	}

	return resourceElasticsearchSecurityUserRead(d, meta)
}

// resourceElasticsearchSecurityUserDelete delete existing user in Elasticsearch
func resourceElasticsearchSecurityUserDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	log.Debugf("User id: %s", id)

	client := meta.(*elastic.Client)
	res, err := client.API.Security.DeleteUser(
		id,
		client.API.Security.DeleteUser.WithContext(context.Background()),
		client.API.Security.DeleteUser.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] User %s not found - removing from state", id)
			log.Warnf("User %s not found - removing from state", id)
			d.SetId("")
			return nil

		}
		return errors.Errorf("Error when delete user %s: %s", id, res.String())
	}

	d.SetId("")

	log.Infof("Deleted user %s successfully", id)
	return nil

}

// Print user object as Json string
func (r *UserSpec) String() string {
	json, _ := json.Marshal(r)
	return string(json)
}

// createUser create or update user in Elasticsearch
func createUser(d *schema.ResourceData, meta interface{}, isUpdate bool) error {
	username := d.Get("username").(string)
	enabled := d.Get("enabled").(bool)
	email := d.Get("email").(string)
	fullName := d.Get("full_name").(string)
	password := d.Get("password").(string)
	passwordHash := d.Get("password_hash").(string)
	roles := convertArrayInterfaceToArrayString(d.Get("roles").(*schema.Set).List())
	metadata := optionalInterfaceJSON(d.Get("metadata").(string))

	user := &UserSpec{
		Enabled:  enabled,
		Email:    email,
		FullName: fullName,
		Roles:    roles,
		Metadata: metadata,
	}

	if isUpdate == false {
		user.Password = password
		user.PasswordHash = passwordHash
	}

	log.Debug("Username: ", username)
	log.Debug("User: ", user)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	client := meta.(*elastic.Client)
	res, err := client.API.Security.PutUser(
		username,
		bytes.NewReader(data),
		client.API.Security.PutUser.WithContext(context.Background()),
		client.API.Security.PutUser.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add user %s: %s", username, res.String())
	}

	return nil
}
