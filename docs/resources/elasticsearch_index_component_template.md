# elasticsearch_index_component_template Resource Source

This resource permit to manage the index component template in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/master/indices-component-template.html

***Supported Elasticsearch version:***
  - v7

## Example Usage

It will create index template.

```tf
resource elasticsearch_index_component_template "test" {
  name 		= "terraform-test"
  template 	= <<EOF
{
	"template": {
		"settings": {
			"index.refresh_interval": "3s"
		},
		"mappings": {
			"_source": {
				"enabled": false
			},
			"properties": {
				"host_name": {
					"type": "keyword"
				},
				"created_at": {
					"type": "date",
					"format": "EEE MMM dd HH:mm:ss Z yyyy"
				}
			}
		}
	}
}
EOF
}
```

## Argument Reference

***The following arguments are supported:***
  - **name**: (required) Identifier for the template.
  - **template**: (required) The template specification. It's a string as JSON object.

## Attribute Reference

NA