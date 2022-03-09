# Elasticsearch Provider

This is a terraform provider that lets you provision elasticsearch resources, compatible with v6, v7 and v8 of elasticsearch.
For Elasticsearch 8, you need to use branch and release 8.x
For Elasticsearch 7, you need to use branch and release 7.x
For Elasticsearch 6, you need to use branch and release 6.x

We fork this project for the following items:
  - use official golang SDK to consume Elasticsearch API: https://github.com/elastic/go-elasticsearch
  - implement importer in terraform
  - migrate to terraform standalone SDK
  - add some resources

## Example Usage

The Elasticsearch provider is used to interact with the
resources supported by Elasticsearch. The provider needs
to be configured with an endpoint URL before it can be used.

***Sample:***
```tf
provider "elasticsearch" {
    urls     = "http://elastic.company.com:9200"
    username = "elastic"
    password = "changeme"
}
```

## Argument Reference

***The following arguments are supported:***
- **urls**: (required) The list of endpoint Elasticsearch URL, separated by comma.
- **username**: (optional) The username to connect on it.
- **password**: (optional) The password to connect on it.
- **insecure**: (optional) To disable the certificate check.
- **cacert_file**: (optional) The CA contend to use if you use custom PKI.
- **retry**: (optional) The number of time you should to retry connexion befaore exist with error. Default to `6`.
- **wait_before_retry**: (optional) The number of time in second we wait before each connexion retry. Default to `10`.


## Resource / Data

- [elasticsearch_index_lifecycle_policy](resources/elasticsearch_index_lifecycle_policy.md)
- [elasticsearch_index_template](resources/elasticsearch_index_template.md)
- [elasticsearch_index_component_template](resources/elasticsearch_index_component_template.md)
- [elasticsearch_index_template_legacy](resources/elasticsearch_index_template_legacy.md)
- [elasticsearch_role](resources/elasticsearch_role.md)
- [elasticsearch_role_mapping](resources/elasticsearch_role_mapping.md)
- [elasticsearch_user](resources/elasticsearch_user.md)
- [elasticsearch_license](resources/elasticsearch_license.md)
- [elasticsearch_snapshot_repository](resources/elasticsearch_snapshot_repository.md)
- [elasticsearch_snapshot_lifecycle_policy](resources/elasticsearch_snapshot_lifecycle_policy.md)
- [elasticsearch_watcher](resources/elasticsearch_watcher.md)
