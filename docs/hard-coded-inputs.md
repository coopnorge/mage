# Hard-coded workflow inputs

Some Mage GitHub Actions workflows need these inputs:

- `oci-image-base`
- `workload-identity-provider`
- `service-account`

If you created your repository using inventory/pallet, the workflow examples in
this repository read these values from GitHub Actions repository variables:

- `vars.PALLET_REGISTRY_URL`
- `vars.PALLET_WORKLOAD_IDENTITY_PROVIDER`
- `vars.PALLET_SERVICE_ACCOUNT`

## What each input does

### `oci-image-base`

`oci-image-base` sets the OCI registry base for image builds.

In the workflows in this repository, Mage reads it from the `OCI_IMAGE_BASE`
environment variable.

Examples in this repository show values such as:

```yaml
oci-image-base: europe-docker.pkg.dev/helloworld-shared-0918
```

### `workload-identity-provider`

The workflow passes `workload-identity-provider` to `google-github-actions/auth`
as `workload_identity_provider` before it logs in to Google Cloud Artifact
Registry.

In the Google Cloud Platform documentation, this term refers to the workload
identity provider name that lets a GitHub repository use a GCP service account
through identity federation.

Examples in this repository show values such as:

```yaml
workload-identity-provider: projects/889992792607/locations/global/workloadIdentityPools/github-actions/providers/github-actions-provider
```

### `service-account`

The workflow passes `service-account` to `google-github-actions/auth` as
`service_account` in the same authentication step.

In the Google Cloud Platform documentation, this term refers to the GCP service
account that the GitHub workflow can use after it authenticates through workload
identity federation.

Examples in this repository show values such as:

```yaml
service-account: gh-ap-helloworld@helloworld-shared-0918.iam.gserviceaccount.com
```

## How inventory/pallet injects them

When you use the workflow example in `docs/index.md`, inventory/pallet-created
repositories read these values from GitHub Actions repository variables:

```yaml
oci-image-base: ${{ vars.PALLET_REGISTRY_URL }}
workload-identity-provider: ${{ vars.PALLET_WORKLOAD_IDENTITY_PROVIDER }}
service-account: ${{ vars.PALLET_SERVICE_ACCOUNT }}
```

The pallet project Crossplane function in `coopnorge/cloud-platform-apis`
creates the repository variables:

- `crossplane-functions/function-palletproject/github_scoped.go`

That code creates GitHub Actions repository variables including:

- `PALLET_REGISTRY_URL`
- `PALLET_SERVICE_ACCOUNT`
- `PALLET_WORKLOAD_IDENTITY_PROVIDER`

Based on the current implementation there, the function derives the values as
follows:

- `PALLET_REGISTRY_URL` comes from
  `europe-docker.pkg.dev/<gcp project id>/<registry name>`
- `PALLET_SERVICE_ACCOUNT` comes from
  `<service-account-name>@<gcp project id>.iam.gserviceaccount.com`
- `PALLET_WORKLOAD_IDENTITY_PROVIDER` comes from
  `projects/<numeric project number>/locations/global/workloadIdentityPools/<pool>/providers/<provider>`

That function generates the exact registry, service account, pool, and provider
names from the project context and may truncate them with a hash suffix when
needed to fit naming limits.

For background on what `workload-identity-provider` and `service-account` mean
and how to set them up for GitHub Actions, see the Inventory docs:

- [Allow GitHub action to access resources in GCP](https://inventory.internal.coop/docs/default/system/cloud-platform/dev_build_deploy/github/guide_github_action_gcp/)
- [Workload identity federation](https://inventory.internal.coop/docs/default/system/cloud-platform/dev_build_deploy/workload_identity_federation/)

## If your project predates inventory/pallet

If you don't already have these values as `vars.PALLET_*`, you can hard-code
them in the workflow as shown in `docs/index.md`.

Places you might already find the values in an older repository include:

- existing GitHub Actions workflow files under `.github/workflows/`
- existing hard-coded Mage workflow inputs such as `oci-image-base`,
  `workload-identity-provider`, or `service-account`
- existing hard-coded renovate workflow inputs such as
  `gcp-workload-identity-provider` or `gcp-service-account`

If your repository uses Helm for deployment, also look at your
`values(-<env>).yaml` files under `<your service>.image`.

For example, you might find an image registry base like:

```yaml
<your service>:
  image: europe-docker.pkg.dev/ecom-integrati-staging-5551
```

In this example, `ecom-integrati-staging-5551` identifies the GCP project ID. If
all environments use the same registry, that project can help you locate the
related workload identity provider and service account in GCP.

If your repository still uses Kustomize instead of Helm, this guide doesn't
document that lookup flow. Please consider migrating to Helm first.

**WARNING** If your repository uses separate registries for different
environments, you need to change this because Mage doesn't support multiple
registries.

### Finding `workload-identity-provider`

In GCP Console, select the project used for the registry and navigate to
`IAM & Admin -> Workload Identity Federation`.

Look for a list of `Workload Identity Pools`. Find a pool named something like
`github-actions`.

Open the pool details. On the right-hand side, look for a list of `Providers`.

Use the provider's edit action and look for the provider resource name under
`Audience`. The value looks like this URL:

```text
https://iam.googleapis.com/projects/625465245830/locations/global/workloadIdentityPools/github-actions/providers/github-actions-provider
```

Use the path part of that URL for `workload-identity-provider`:

```text
projects/625465245830/locations/global/workloadIdentityPools/github-actions/providers/github-actions-provider
```

### Finding `service-account`

In the same pool details view, the provider list also has a tab called
`Connected service accounts`.

That tab lists the service accounts connected to the pool. Choose the service
account you want to use for the `service-account` input.

The UI may not show the full email address directly there, but you can compose
it as:

```text
<service-account-name>@<gcp project>.iam.gserviceaccount.com
```

For example:

```text
gh-action-pipeline-account@ecom-integrati-staging-5551.iam.gserviceaccount.com
```
