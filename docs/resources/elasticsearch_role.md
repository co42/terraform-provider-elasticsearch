# elasticsearch_role Resource Source

This resource permit to manage role in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role.html

***Supported Elasticsearch version:***
  - v6
  - v7

## Example Usage

It will create role called `terraform-test` with some privileges.

```tf
resource elasticsearch_role "test" {
  name = "terraform-test"
  indices {
	  names = ["logstash-*"]
	  privileges = ["read"]
  }
  indices {
	  names = ["logstash-*"]
	  privileges = ["read2"]
  }
  cluster = ["all"]
}
```

## Argument Reference

***The following arguments are supported:***
  - **name**: (required) The role name to create
  - **cluster**: (optional) A list of cluster privileges. These privileges define the cluster level actions that users with this role are able to execute.
  - **run_as**: (optional) A list of users that the owners of this role can impersonate.
  - **global**: (optional) A string as JSON object defining global privileges. A global privilege is a form of cluster privilege that is request-aware. Support for global privileges is currently limited to the management of application privileges.
  - **metadata**: (optional) A string as JSON object meta-data. Within the metadata object, keys that begin with _ are reserved for system usage.
  - **indices**: (optional) A list of indices permissions entries. Look the indice object below.
  - **applications**: (optional) A list of application privilege entries. Look the application object below.


***Indice object***:
  - **names**: (required) A list of indices (or index name patterns) to which the permissions in this entry apply.
  - **privileges**: (required) A list of The index level privileges that the owners of the role have on the specified indices.
  - **query**: (optional) A search query that defines the documents the owners of the role have read access to. A document within the specified indices must match this query in order for it to be accessible by the owners of the role. It's a string or a string as JSON object.
  - **field_security**: (optional) The document fields that the owners of the role have read access to. It's a string as JSON object

***Application object***:
  - **application**: (required) The name of the application to which this entry applies.
  - **privileges**: (optional)  A list of strings, where each element is the name of an application privilege or action.
  - **resources**: (optional) A list resources to which the privileges are applied.

## Attribute Reference

NA