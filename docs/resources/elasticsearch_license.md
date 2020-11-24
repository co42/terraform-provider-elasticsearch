# elasticsearch_license Resource Source

This resource permit to manage license in Elasticsearch.
You can use enterprise license file or enable basic license.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/update-license.html

***Supported Elasticsearch version:***
  - v6
  - v7

## Example Usage

It will enabled basic license.

```tf
resource elasticsearch_license "test" {
  use_basic_license = "true"
}
```

## Argument Reference

***The following arguments are supported:***
  - **license**: (optional) The license contend file.
  - **use_basic_license**: (required) Set `true` to use basic licence.

## Attribute Reference

NA