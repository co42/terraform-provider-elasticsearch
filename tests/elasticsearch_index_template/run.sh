#!/bin/sh

cat <<EOT > ${HOME}/.terraformrc
provider_installation {
    filesystem_mirror {
        path    = "${PWD}/../../registry"
        include = ["registry.terraform.io/co42/elasticsearch"]
    }
    direct {
        exclude = ["registry.terraform.io/co42/elasticsearch"]
    }
}
EOT

rm -rf .terraform*
terraform init
terraform apply