# terraform-provider-elasticsearch



This is a terraform provider that lets you provision elasticsearch resources, compatible with v6 and v7 of elasticsearch.

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

The Elasticsearch provider is used to interact with the
resources supported by Elasticsearch. The provider needs
to be configured with an endpoint URL before it can be used.

__Sample:__
```tf
provider "elasticsearch" {
    urls     = "http://elastic.company.com:9200"
    username = "elastic"
    password = "changeme"
}
```

__The following arguments are supported:__
- **urls**: (required) The list of endpoint Elasticsearch URL, separated by comma.
- **username**: (optional) The username to connect on it.
- **password**: (optional) The password to connect on it.
- **insecure**: (optional) To disable the certificate check.
- **cacert_file**: (optional) The CA contend to use if you use custom PKI.

### Resources

#### Role

This resource permit to manage role in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role.html

__Supported Elasticsearch version:__
  - v6
  - v7

__Sample:__
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

__The following arguments are supported:__
  - **name**: (required) The role name to create
  - **cluster**: (optional) A list of cluster privileges. These privileges define the cluster level actions that users with this role are able to execute.
  - **run_as**: (optional) A list of users that the owners of this role can impersonate.
  - **global**: (optional) A string as JSON object defining global privileges. A global privilege is a form of cluster privilege that is request-aware. Support for global privileges is currently limited to the management of application privileges.
  - **metadata**: (optional) A string as JSON object meta-data. Within the metadata object, keys that begin with _ are reserved for system usage.
  - **indices**: (optional) A list of indices permissions entries. Look the indice object below.
  - **applications**: (optional) A list of application privilege entries. Look the application object below.


__Indice object__:
  - **names**: (required) A list of indices (or index name patterns) to which the permissions in this entry apply.
  - **privileges**: (required) A list of The index level privileges that the owners of the role have on the specified indices.
  - **query**: (optional) A search query that defines the documents the owners of the role have read access to. A document within the specified indices must match this query in order for it to be accessible by the owners of the role. It's a string or a string as JSON object.
  - **field_security**: (optional) The document fields that the owners of the role have read access to. It's a string as JSON object

__Application object__:
  - **application**: (required) The name of the application to which this entry applies.
  - **privileges**: (optional)  A list of strings, where each element is the name of an application privilege or action.
  - **resources**: (optional) A list resources to which the privileges are applied.


#### Role mapping

This resource permit to manage role mapping ins Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html

__Supported Elasticsearch version:__
  - v6
  - v7

__Sample__:
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

__The following arguments are supported:__
  - **name:** (required) The distinct name that identifies the role mapping.
  - **enabled:** (optional) Mappings that have enabled set to false are ignored when role mapping is performed.
  - **rules**: (required) The rules that determine which users should be matched by the mapping. A rule is a logical condition that is expressed by using a JSON DSL. It's a string as JSON object.
  - **roles**: (required) A list of role names that are granted to the users that match the role mapping rules.
  - **metadata:** (optional) Additional metadata that helps define which roles are assigned to each user. It's a string as JSON object.


#### User

This resource permit to manage internal user in Elasticsearch.
You can see the API documenation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-user.html

__Supported Elasticsearch version:__
  - v6
  - v7

__Sample:__
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

__The following arguments are supported:__
  - **username**: (required) An identifier for the user.
  - **email**: (required) The email of the user.
  - **full_name**: (optional) The full name of the user.
  - **password**: (optional) The user’s password. Passwords must be at least 6 characters long. When adding a user, one of password or password_hash is required.
  - **password_hash**: (optional) A hash of the user’s password
  - **enabled**: (optional) Specifies whether the user is enabled
  - **roles**: (required) A set of roles the user has
  - **metadata**: (optional) Arbitrary metadata that you want to associate with the user

#### Index lifecycle policy

This resource permit to manage the index lifecycle policy in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-put-lifecycle.html

__Supported Elasticsearch version:__
  - v6
  - v7

__Sample:__
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

__The following arguments are supported:__
  - **name**: (required) Identifier for the policy.
  - **policy**: (required) The policy specification. It's a string as JSON object.

#### Index template

This resource permit to manage the index template in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-templates.html

__Supported Elasticsearch version:__
  - v6
  - v7

__Sample:__
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

__The following arguments are supported:__
  - **name**: (required) Identifier for the template.
  - **template**: (required) The template specification. It's a string as JSON object.

#### License

This resource permit to manage license in Elasticsearch.
You can use enterprise license file or enable basic license.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/update-license.html

__Supported Elasticsearch version:__
  - v6
  - v7

__Sample:__
```tf
resource "elasticsearch_license" "test" {
  use_basic_license = "true"
}
```

__The following arguments are supported:__
  - **license**: (optional) The license contend file.
  - **use_basic_license**: (required) Set `true` to use basic licence.

#### Snapshot repository

This resource permit to manage snapshot repository in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-snapshots.html

__Supported Elasticsearch version:__
  - v6
  - v7

__Sample:__
```tf
resource "elasticsearch_snapshot_repository" "test" {
  name		= "terraform-test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
  }
}
```

__The following arguments are supported:__
  - **name**: (required) Identifier for the repository.
  - **type**: (required) The repository type.
  - **settings**: (required) The list of settings. It's a map of string.

#### Elasticsearch snapshot lifecycle policy

This resource permit to manage snapshot lifecyle policy.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/slm-api-put.html

__Supported Elasticsearch version:__
  - v7

__Sample:__
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

__The following arguments are supported:__
  - **name**: (required) Identifier for the policy.
  - **snapshot_name**: (required) A name automatically given to each snapshot performed by this policy.
  - **schedule**: (required) A periodic or absolute time schedule.
  - **repository**: (required) The snapshot repository that will contain snapshots created by this policy.
  - **configs**: (optional) Configuration for each snapshot that will be created by this policy. It's a string as JSON object.


#### Watcher

This resource permit to manage watcher in Elasticsearch
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-put-watch.html

__Supported Elasticsearch version:__
  - v6
  - v7

__Sample:__
```tf
resource "elasticsearch_watcher" "test" {
  name		= "terraform-test"
  trigger	= <<EOF
{
	"schedule" : { "cron" : "0 0/1 * * * ?" }
}
EOF
  input		= <<EOF
{
	"search" : {
      "request" : {
        "indices" : [
          "logstash*"
        ],
        "body" : {
          "query" : {
            "bool" : {
              "must" : {
                "match": {
                   "response": 404
                }
              },
              "filter" : {
                "range": {
                  "@timestamp": {
                    "from": "{{ctx.trigger.scheduled_time}}||-5m",
                    "to": "{{ctx.trigger.triggered_time}}"
                  }
                }
              }
            }
          }
        }
      }
    }
}
EOF
  condition		= <<EOF
{
	"compare" : { "ctx.payload.hits.total" : { "gt" : 0 }}
}
EOF
  actions		= <<EOF
{
	"email_admin" : {
      "email" : {
        "to" : "admin@domain.host.com",
        "subject" : "404 recently encountered"
      }
    }
}
EOF
}
```

__The following arguments are supported:__
  - **name**: (required) Identifier for the watcher.
  - **trigger**: (optional) The trigger that defines when the watch should run. It's a string as JSON object.
  - **input**: (optional) The input that defines the input that loads the data for the watch. It's a string as JSON object.
  - **condition**: (optional) The condition that defines if the actions should be run. It's a string as JSON object.
  - **actions**: (optional) The list of actions that will be run if the condition matches. It's a string as JSOn object.
  - **throttle_period**: (optional) The minimum time between actions being run.
  - **metadata**: (optional) Metadata json that will be copied into the history entries. It's a string as JSON object.

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
