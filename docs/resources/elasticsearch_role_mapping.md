# elasticsearch_role_mapping Resource Source

This resource permit to manage role mapping ins Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html

***Supported Elasticsearch version:***
  - v6
  - v7

## Example Usage

It will map LDAP group with `superuser` role.

```tf
resource elasticsearch_role_mapping "test" {
  name = "terraform-test"
  enabled = "true"
  roles = ["superuser"]
  rules = <<EOF
{
	"field": {
		"groups": "cn=admins,dc=example,dc=com"
	}
}
EOF
}
```

## Argument Reference

***The following arguments are supported:***
  - **name:** (required) The distinct name that identifies the role mapping.
  - **enabled:** (optional) Mappings that have enabled set to false are ignored when role mapping is performed.
  - **rules**: (required) The rules that determine which users should be matched by the mapping. A rule is a logical condition that is expressed by using a JSON DSL. It's a string as JSON object.
  - **roles**: (required) A list of role names that are granted to the users that match the role mapping rules.
  - **metadata:** (optional) Additional metadata that helps define which roles are assigned to each user. It's a string as JSON object.


## Attribute Reference

NA