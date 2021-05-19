package es

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
)

// diffSuppressIndexTemplate permit to compare template in current state vs from API
func diffSuppressIndexTemplate(k, old, new string, d *schema.ResourceData) bool {
	var oo, no map[string]interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	// Go under first level to compare
	for key, data := range no {
		no = data
		break
	}

	// Add default parameters on new index template if needed
	if _, ok := no["order"]; !ok {
		no["order"] = 0
	}
	if _, ok := no["settings"]; !ok {
		no["settings"] = make(map[string]interface{})
	}
	if _, ok := no["mappings"]; !ok {
		no["mappings"] = make(map[string]interface{})
	}
	if _, ok := no["aliases"]; !ok {
		no["aliases"] = make(map[string]interface{})
	}

	return reflect.DeepEqual(oo[d.Id()], parseAllDotProperties(no))
}

// suppressEquivalentJSON permit to compare state store as JSON string
func suppressEquivalentJSON(k, old, new string, d *schema.ResourceData) bool {
	var oldObj, newObj interface{}
	if err := json.Unmarshal([]byte(old), &oldObj); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &newObj); err != nil {
		return false
	}
	return reflect.DeepEqual(oldObj, newObj)
}

// suppressLicense permit to compare license in current state VS API
func suppressLicense(k, old, new string, d *schema.ResourceData) bool {

	oldObj := &LicenseSpec{}
	newObjTemp := make(License)
	if err := json.Unmarshal([]byte(old), oldObj); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &newObjTemp); err != nil {
		return false
	}
	newObj := newObjTemp["license"]

	newObj.Signature = ""
	oldObj.Signature = ""

	log.Debugf("Old: %s\nNew: %s", oldObj, newObj)

	return reflect.DeepEqual(oldObj, newObj)
}

// parseAllDotProperties permit to convert elasticsearch attributes with dot in sub structure
func parseAllDotProperties(data map[string]interface{}) map[string]interface{} {

	result := make(map[string]interface{})
	for k, v := range data {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Map {
			parseDotPropertie(k, parseAllDotProperties(v.(map[string]interface{})), result)
		} else {
			parseDotPropertie(k, v, result)
		}
	}

	return result
}

// parseDotPropertie handle the recursivity to transform attribute that contain dot in sub structure
func parseDotPropertie(key string, value interface{}, result map[string]interface{}) {
	if strings.Contains(key, ".") {
		listKey := strings.Split(key, ".")
		if _, ok := result[listKey[0]]; !ok {
			result[listKey[0]] = make(map[string]interface{})
		}
		parseDotPropertie(strings.Join(listKey[1:], "."), value, result[listKey[0]].(map[string]interface{}))
	} else {
		result[key] = value
	}
}
