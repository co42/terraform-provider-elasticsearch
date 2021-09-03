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


resource "elasticsearch_index_template" "test" {
    name = "test"
    template = jsondecode(jsonencode(file("test.json")))
}