package es

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func diffSuppressIndexTemplate(k, old, new string, d *schema.ResourceData) bool {
	var oo, no map[string]interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
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

func suppressEquivalentJson(k, old, new string, d *schema.ResourceData) bool {
	var oldObj, newObj interface{}
	if err := json.Unmarshal([]byte(old), &oldObj); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &newObj); err != nil {
		return false
	}
	return reflect.DeepEqual(oldObj, newObj)
}

func suppressLicense(k, old, new string, d *schema.ResourceData) bool {
	var oldObj, newObj map[string]interface{}
	if err := json.Unmarshal([]byte(old), &oldObj); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &newObj); err != nil {
		return false
	}

	// Remove field status to compare
	delete(oldObj["license"].(map[string]interface{}), "status")
	delete(oldObj["license"].(map[string]interface{}), "issue_date")
	delete(oldObj["license"].(map[string]interface{}), "expiry_date")

	delete(newObj["license"].(map[string]interface{}), "signature")

	return reflect.DeepEqual(oldObj, newObj)
}

func diffSuppressIndexLifecyclePolicy(k, old, new string, d *schema.ResourceData) bool {
	var oo, no map[string]interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	cleanOo := map[string]interface{}{
		"policy": oo[d.Id()].(map[string]interface{})["policy"],
	}

	return reflect.DeepEqual(cleanOo, no)
}

func diffSuppressIngestPipeline(k, old, new string, d *schema.ResourceData) bool {
	var oo, no interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	return reflect.DeepEqual(oo, no)
}

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
