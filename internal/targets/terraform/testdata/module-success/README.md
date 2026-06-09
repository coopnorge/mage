<!-- BEGIN_TF_DOCS -->
# Terraform module data-platform domain

This module sets up infrastructure for a data-platform domain

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

No modules.

## Resources

No resources.

## Developing

Read more about developing an terraform module in the [playbook][playbook-tf-dev]

## Versioning

Read more about versioning and publishing in the [playbook][playbook-tf-version]

[playbook-tf-dev]: https://inventory.internal.coop/docs/default/component/guidelines/languages/terraform/#terraform
[playbook-tf-version]: https://inventory.internal.coop/docs/default/component/guidelines/languages/terraform/#versioning
<!-- END_TF_DOCS -->