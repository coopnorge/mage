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


resource "random_id" "server" {
  byte_lengtha = 8
}

