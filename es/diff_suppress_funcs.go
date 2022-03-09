package es

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/elastic/go-ucfg"
	"github.com/elastic/go-ucfg/diff"
	ucfgjson "github.com/elastic/go-ucfg/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

// diffSuppressIndexTemplateLegacy permit to compare template in current state vs from API
func diffSuppressIndexTemplateLegacy(k, old, new string, d *schema.ResourceData) bool {

	oo := &elastic.IndicesGetTemplateResponse{}
	no := &elastic.IndicesGetTemplateResponse{}

	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		fmt.Printf("[ERR] Error when converting to IndicesGetComponentTemplate: %s", err.Error())
		log.Errorf("Error when converting to IndicesGetComponentTemplate: %s", err.Error())
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		fmt.Printf("[ERR] Error when converting to IndicesGetComponentTemplate: %s", err.Error())
		log.Errorf("Error when converting to IndicesGetComponentTemplate: %s", err.Error())
		return false
	}

	// inits default values

	if oo.Aliases == nil {
		oo.Aliases = make(map[string]interface{})
	}
	if oo.Mappings == nil {
		oo.Mappings = make(map[string]interface{})
	}
	if oo.Settings == nil {
		oo.Settings = make(map[string]interface{})
	}

	if no.Aliases == nil {
		no.Aliases = make(map[string]interface{})
	}
	if no.Mappings == nil {
		no.Mappings = make(map[string]interface{})
	}
	if no.Settings == nil {
		no.Settings = make(map[string]interface{})
	}

	// force undot properties to compare the same think
	oo.Aliases = parseAllDotProperties(oo.Aliases)
	oo.Mappings = parseAllDotProperties(oo.Mappings)
	oo.Settings = parseAllDotProperties(oo.Settings)

	no.Aliases = parseAllDotProperties(no.Aliases)
	no.Mappings = parseAllDotProperties(no.Mappings)
	no.Settings = parseAllDotProperties(no.Settings)

	return reflect.DeepEqual(oo, no)
}

// suppressEquivalentJSON permit to compare state store as JSON string
func suppressEquivalentJSON(k, old, new string, d *schema.ResourceData) bool {

	if old == "" {
		old = `{}`
	}
	if new == "" {
		new = `{}`
	}
	confOld, err := ucfgjson.NewConfig([]byte(old), ucfg.PathSep("."))
	if err != nil {
		fmt.Printf("[ERR] Error when converting current Json: %s\ndata: %s", err.Error(), old)
		log.Errorf("Error when converting current Json: %s\ndata: %s", err.Error(), old)
		return false
	}
	confNew, err := ucfgjson.NewConfig([]byte(new), ucfg.PathSep("."))
	if err != nil {
		fmt.Printf("[ERR] Error when converting new Json: %s\ndata: %s", err.Error(), new)
		log.Errorf("Error when converting new Json: %s\ndata: %s", err.Error(), new)
		return false
	}

	currentDiff := diff.CompareConfigs(confOld, confNew)
	log.Debugf("Diff\n: %s", currentDiff.GoStringer())

	return !currentDiff.HasChanged()
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

	// Fix `limit` filed is string, not number on ES response
	converFields := []string{
		"limit",
		"number_of_routing_shards",
		"number_of_replicas",
		"number_of_shards",
	}
	for _, field := range converFields {
		if key == field {
			if reflect.ValueOf(value).Kind() == reflect.Float64 {
				result[key] = strconv.Itoa(int(value.(float64)))
			}
			break
		}
	}

}

// diffSuppressIndexComponentTemplate permit to compare index component template in current state vs from API
func diffSuppressIndexComponentTemplate(k, old, new string, d *schema.ResourceData) bool {
	oo := &elastic.IndicesGetComponentTemplate{}
	no := &elastic.IndicesGetComponentTemplate{}

	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		fmt.Printf("[ERR] Error when converting to IndicesGetComponentTemplate: %s", err.Error())
		log.Errorf("Error when converting to IndicesGetComponentTemplate: %s", err.Error())
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		fmt.Printf("[ERR] Error when converting to IndicesGetComponentTemplate: %s", err.Error())
		log.Errorf("Error when converting to IndicesGetComponentTemplate: %s", err.Error())
		return false
	}

	// inits default values
	if oo.Template != nil {
		if oo.Template.Aliases == nil {
			oo.Template.Aliases = make(map[string]interface{})
		}
		if oo.Template.Mappings == nil {
			oo.Template.Mappings = make(map[string]interface{})
		}
		if oo.Template.Settings == nil {
			oo.Template.Settings = make(map[string]interface{})
		}
	}
	if no.Template != nil {
		if no.Template.Aliases == nil {
			no.Template.Aliases = make(map[string]interface{})
		}
		if no.Template.Mappings == nil {
			no.Template.Mappings = make(map[string]interface{})
		}
		if no.Template.Settings == nil {
			no.Template.Settings = make(map[string]interface{})
		}
	}

	// force undot properties to compare the same think
	if oo.Template != nil {
		oo.Template.Aliases = parseAllDotProperties(oo.Template.Aliases)
		oo.Template.Mappings = parseAllDotProperties(oo.Template.Mappings)
		oo.Template.Settings = parseAllDotProperties(oo.Template.Settings)
	}
	if no.Template != nil {
		no.Template.Aliases = parseAllDotProperties(no.Template.Aliases)
		no.Template.Mappings = parseAllDotProperties(no.Template.Mappings)
		no.Template.Settings = parseAllDotProperties(no.Template.Settings)
	}

	return reflect.DeepEqual(oo, no)
}

// diffSuppressIndexTemplate permit to compare index template in current state vs from API
func diffSuppressIndexTemplate(k, old, new string, d *schema.ResourceData) bool {
	oo := &elastic.IndicesGetIndexTemplate{}
	no := &elastic.IndicesGetIndexTemplate{}

	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		fmt.Printf("[ERR] Error when converting to IndicesGetIndexTemplate on old object: %s", err.Error())
		log.Errorf("Error when converting to IndicesGetIndexTemplate on old object: %s\n%s", err.Error(), old)
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		fmt.Printf("[ERR] Error when converting to IndicesGetIndexTemplate on new object: %s", err.Error())
		log.Errorf("Error when converting to IndicesGetIndexTemplate on new object: %s\n%s", err.Error(), new)
		return false
	}

	// inits default values
	if oo.Template != nil {
		if oo.Template.Aliases == nil {
			oo.Template.Aliases = make(map[string]interface{})
		}
		if oo.Template.Mappings == nil {
			oo.Template.Mappings = make(map[string]interface{})
		}
		if oo.Template.Settings == nil {
			oo.Template.Settings = make(map[string]interface{})
		}
	}
	if no.Template != nil {
		if no.Template.Aliases == nil {
			no.Template.Aliases = make(map[string]interface{})
		}
		if no.Template.Mappings == nil {
			no.Template.Mappings = make(map[string]interface{})
		}
		if no.Template.Settings == nil {
			no.Template.Settings = make(map[string]interface{})
		}
	}

	// force undot properties to compare the same think
	if oo.Template != nil {
		oo.Template.Aliases = parseAllDotProperties(oo.Template.Aliases)
		oo.Template.Mappings = parseAllDotProperties(oo.Template.Mappings)
		oo.Template.Settings = parseAllDotProperties(oo.Template.Settings)
	}
	if no.Template != nil {
		no.Template.Aliases = parseAllDotProperties(no.Template.Aliases)
		no.Template.Mappings = parseAllDotProperties(no.Template.Mappings)
		no.Template.Settings = parseAllDotProperties(no.Template.Settings)
	}

	return reflect.DeepEqual(no, oo)
}


// diffSuppressTransform permit to compare transform in current state vs from API
func diffSuppressTransform(k, old, new string, d *schema.ResourceData) bool {
	oo := &Transform{}
	no := &Transform{}

	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		fmt.Printf("[ERR] Error when converting to Transform on old object: %s", err.Error())
		log.Errorf("Error when converting to Transform on old object: %s\n%s", err.Error(), old)
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		fmt.Printf("[ERR] Error when converting to Transform on new object: %s", err.Error())
		log.Errorf("Error when converting to Transform on new object: %s\n%s", err.Error(), new)
		return false
	}

	oo.Id = ""
	oo.CreateTime = 0
	oo.Version = ""


// diffSuppressIngestPipeline permit to compare ingest pipeline in current state vs from API
func diffSuppressIngestPipeline(k, old, new string, d *schema.ResourceData) bool {
	oo := &elastic.IngestGetPipeline{}
	no := &elastic.IngestGetPipeline{}

	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		fmt.Printf("[ERR] Error when converting to IngestGetPipeline on old object: %s", err.Error())
		log.Errorf("Error when converting to IngestGetPipeline on old object: %s\n%s", err.Error(), old)
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		fmt.Printf("[ERR] Error when converting to IngestGetPipeline on new object: %s", err.Error())
		log.Errorf("Error when converting to IngestGetPipeline on new object: %s\n%s", err.Error(), new)
		return false
	}

	return reflect.DeepEqual(no, oo)
}
