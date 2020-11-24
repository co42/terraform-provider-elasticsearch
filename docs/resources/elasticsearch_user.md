# elasticsearch_user Resource Source

This resource permit to manage internal user in Elasticsearch.
You can see the API documenation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-user.html

***Supported Elasticsearch version:***
  - v6
  - v7

## Example Usage

It will create new user called `terraform-test` with some roles.

```tf
resource elasticsearch_user "test" {
  username 	= "terraform-test"
  enabled 	= "true"
  email 	= "no@no.no"
  full_name = "test"
  password 	= "changeme"
  roles 	= ["kibana_user"]
}
```

## Argument Reference

***The following arguments are supported:***
  - **username**: (required) An identifier for the user.
  - **email**: (required) The email of the user.
  - **full_name**: (optional) The full name of the user.
  - **password**: (optional) The user’s password. Passwords must be at least 6 characters long. When adding a user, one of password or password_hash is required.
  - **password_hash**: (optional) A hash of the user’s password
  - **enabled**: (optional) Specifies whether the user is enabled
  - **roles**: (required) A set of roles the user has
  - **metadata**: (optional) Arbitrary metadata that you want to associate with the user

## Attribute Reference

NA