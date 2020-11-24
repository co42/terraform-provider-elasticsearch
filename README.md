# terraform-provider-elasticsearch

[![CircleCI](https://circleci.com/gh/disaster37/terraform-provider-elasticsearch/tree/7.x.svg?style=svg)](https://circleci.com/gh/disaster37/terraform-provider-elasticsearch/tree/7.x)
[![Go Report Card](https://goreportcard.com/badge/github.com/disaster37/terraform-provider-elasticsearch)](https://goreportcard.com/report/github.com/disaster37/terraform-provider-elasticsearch)
[![GoDoc](https://godoc.org/github.com/disaster37/terraform-provider-elasticsearch?status.svg)](http://godoc.org/github.com/disaster37/terraform-provider-elasticsearch)
[![codecov](https://codecov.io/gh/disaster37/terraform-provider-elasticsearch/branch/7.x/graph/badge.svg)](https://codecov.io/gh/disaster37/terraform-provider-elasticsearch/branch/7.x)

This is a terraform provider that lets you provision elasticsearch resources, compatible with v6 and v7 of elasticsearch.
For Elasticsearch 7, you need to use branch and release 7.x
For Elasticsearch 6, you need to use branch and release 6.x

We fork this project for the following items:
  - use official golang SDK to consume Elasticsearch API: https://github.com/elastic/go-elasticsearch
  - implement importer in terraform
  - migrate to terraform standalone SDK
  - add some resources

## Installation

[Go to terraform registry](https://registry.terraform.io/providers/disaster37/elasticsearch/latest)

## Documentation

[Read provider documentation](docs/index.md)


## Development

### Requirements

* [Golang](https://golang.org/dl/) >= 1.11
* [Terrafrom](https://www.terraform.io/) >= 0.12


```
go build -o /path/to/binary/terraform-provider-elasticsearch
```

## Licence

See LICENSE.

## Contributing

1. Fork it ( https://github.com/disaster37/terraform-provider-elasticsearch/fork )
2. Go to the right branch (7.x for Elasticsearch 7 or 6.x for Elasticsearch 6) (`git checkout 7.x`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Add feature, add acceptance test and tets your code (`ELASTICSEARCH_URLS=http://127.0.0.1:9200 ELASTICSEARCH_USERNAME=elastic ELASTICSEARCH_PASSWORD=changeme make testacc`)
5. Commit your changes (`git commit -am 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create a new Pull Request

