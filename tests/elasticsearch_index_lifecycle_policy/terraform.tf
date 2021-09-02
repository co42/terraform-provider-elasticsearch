terraform {
  required_providers {
    elasticsearch = {
      source = "disaster37/elasticsearch"
      version = "1.0.0"
    }
  }

  
}

provider "elasticsearch" {
    urls      = "http://elasticsearch:9200"
    username  = "elastic"
    password  = "changeme"
}


resource "elasticsearch_index_lifecycle_policy" "test" {
    name = "test"
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