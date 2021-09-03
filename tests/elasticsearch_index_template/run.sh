#!/bin/sh

cat <<EOT > ${HOME}/.terraformrc
provider_installation {
    filesystem_mirror {
        path    = "${PWD}/../../registry"
        include = ["registry.terraform.io/disaster37/elasticsearch"]
    }
    direct {
        exclude = ["registry.terraform.io/disaster37/elasticsearch"]
    }
}
EOT

rm -rf .terraform*
terraform init
terraform apply