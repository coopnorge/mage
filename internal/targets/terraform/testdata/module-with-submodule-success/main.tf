terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }

  required_version = "~> 1.0"
}

provider "random" {
}

module "sub1" {
  source = "./submodule"
  var1   = "var1"
}
