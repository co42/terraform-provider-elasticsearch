# elasticsearch_snapshot_repository Resource Source

This resource permit to manage snapshot repository in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-snapshots.html

***Supported Elasticsearch version:***
  - v6
  - v7

## Example Usage

It will create snapshot repository.

```tf
resource elasticsearch_snapshot_repository "test" {
  name		= "terraform-test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
  }
}
```

## Argument Reference

***The following arguments are supported:***
  - **name**: (required) Identifier for the repository.
  - **type**: (required) The repository type.
  - **settings**: (required) The list of settings. It's a map of string.

## Attribute Reference

NA