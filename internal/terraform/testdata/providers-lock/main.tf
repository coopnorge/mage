terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "< 3.8.2"
    }
  }

  required_version = "~> 1.0"
}

provider "random" {
}
