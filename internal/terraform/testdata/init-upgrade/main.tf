terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "< 3.7.3"
    }
  }

  required_version = "~> 1.0"
}

provider "random" {
}
