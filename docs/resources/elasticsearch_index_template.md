# elasticsearch_index_template Resource Source

This resource permit to manage the index template in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/index-templates.html

***Supported Elasticsearch version:***
  - v7

## Example Usage

It will create index template.

```tf
resource elasticsearch_index_template "test" {
  name 		= "terraform-test"
  template 	= <<EOF
{
	"index_patterns": ["test-index-template"],
	"template": {
		"settings": {
			"index.refresh_interval": "5s",
			"index.lifecycle.name": "policy-logstash-backup",
    		"index.lifecycle.rollover_alias": "logstash-backup-alias"
		}
	},
	"priority": 2
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