<!-- BEGIN_TF_DOCS -->
# Terraform module test example

This module is a test example

## Example usage

```hcl
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
```
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | ~> 1.0 |
| <a name="requirement_random"></a> [random](#requirement\_random) | 3.7.2 |

## Inputs

No inputs.

## Outputs

No outputs.

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_sub1"></a> [sub1](#module\_sub1) | ./submodule | n/a |

## Resources

No resources.

## Developing

Read more about developing a terraform module in the [playbook][playbook-tf-dev]

## Versioning

Read more about versioning and publishing in the [playbook][playbook-tf-version]

[playbook-tf-dev]: https://inventory.internal.coop/docs/default/component/guidelines/languages/terraform/#terraform
[playbook-tf-version]: https://inventory.internal.coop/docs/default/component/guidelines/languages/terraform/#versioning
<!-- END_TF_DOCS -->