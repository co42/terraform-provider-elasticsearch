# terraform-provider-elasticsearch

[![Build Status](https://travis-ci.org/phillbaker/terraform-provider-elasticsearch.svg?branch=master)](https://travis-ci.org/phillbaker/terraform-provider-elasticsearch)

This is a terraform provider that lets you provision elasticsearch resources, compatible with v5, v6 and v7 of elasticsearch. Based off of an [original PR to Terraform](https://github.com/hashicorp/terraform/pull/13238).

## Installation

[Download a binary](https://github.com/phillbaker/terraform-provider-elasticsearch/releases), and put it in a good spot on your system. Then update your `~/.terraformrc` to refer to the binary:

```hcl
providers {
  elasticsearch = "/path/to/terraform-provider-elasticsearch"
}
```

See [the docs for more information](https://www.terraform.io/docs/plugins/basics.html).

## Usage

### Provider

```tf
provider "elasticsearch" {
    urls     = "http://elastic.company.com:9200"
    username = "elastic"
    password = "changeme"
}
```

### Resources

#### Elasticsearch role

Supported Elasticsearch version:
  - v6
  - v7

```tf
resource "elasticsearch_role" "test" {
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

#### Elasticsearch role mapping

Supported Elasticsearch version:
  - v6
  - v7

```tf
resource "elasticsearch_role_mapping" "test" {
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

#### Elasticsearch user

Supported Elasticsearch version:
  - v6
  - v7

```tf
resource "elasticsearch_user" "test" {
  username 	= "terraform-test"
  enabled 	= "true"
  email 	= "no@no.no"
  full_name = "test"
  password 	= "changeme"
  roles 	= ["kibana_user"]
}
```

#### Elasticsearch lifecycle policy

Supported Elasticsearch version:
  - v6
  - v7

```tf
resource "elasticsearch_index_lifecycle_policy" "test" {
  name = "terraform-test"
  policy = <<EOF
{
  "policy": {
    "phases": {
      "warm": {
        "min_age": "10d",
        "actions": {
          "forcemerge": {
            "max_num_segments": 1
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
EOF
}
```

#### Elasticsearch index template

Supported Elasticsearch version:
  - v6
  - v7

```tf
resource "elasticsearch_index_template" "test" {
  name 		= "terraform-test"
  template 	= <<EOF
{
  "index_patterns": [
    "test"
  ],
  "settings": {
    "index.refresh_interval": "5s",
	"index.lifecycle.name": "policy-logstash-backup",
    "index.lifecycle.rollover_alias": "logstash-backup-alias"
  },
  "order": 2
}
EOF
}
```

#### Elasticsearch license

Supported Elasticsearch version:
  - v6
  - v7

```tf
resource "elasticsearch_license" "test" {
  use_basic_license = "true"
}
```

#### Elasticsearch snapshot repository

```tf
resource "elasticsearch_snapshot_repository" "test" {
  name		= "terraform-test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
  }
}
```

#### Elasticsearch snapshot lifecycle policy

Supported Elasticsearch version:
  - v7

```tf
resource "elasticsearch_snapshot_lifecycle_policy" "test" {
  name			= "terraform-test"
  snapshot_name = "<daily-snap-{now/d}>"
  schedule 		= "0 30 1 * * ?"
  repository    = "${elasticsearch_snapshot_repository.test.name}"
  configs		= <<EOF
{
	"indices": ["test-*"],
	"ignore_unavailable": false,
	"include_global_state": false
}
EOF
}
```

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
2. Go to develop branch (`git checkout develop`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Add feature, add acceptance test and tets your code (`make testacc`)
5. Commit your changes (`git commit -am 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create a new Pull Request
