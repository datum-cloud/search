# API Design

The Search service provides APIs for searching resources and managing the
resources available to search. Platform operators can quickly use this service
to search for resources they need across the platform.

## Motivation

The platform supports quick and easy registration of resources though Custom
Resource Definitions and API aggregation. Given this, it's important our search
service can adjust dynamically at runtime to resources being indexed and
searchable so it doesn't require software changes to index new resources.

## Goals

- Introduce a dynamic method of configuring which resources in the platform
  should be indexed and searchable
- Enable platform operators to quickly search across resource types while also
  allowing field based selection to narrow the scope of the search
- Support full-text search capabilities configurable through dynamic index
  policies


## Proposal

The Search API will initially offer these new API resources to end-users:

- **ResourceIndexPolicy** configures which resources in the platform should be
  indexed and which fields within the resource are applicable to indexing
- **ResourceSearchQuery** allows users to execute field filtering and full-text
  searching capabilities across all indexed resources


> [!IMPORTANT]
>
> The system does not support AuthZ based policies to ensure users have access
> to results returned by search queries. This API should only be used for
> non-sensitive resources and for internal-use only.
>
> This service will be made multi-tenant friendly in the future.

## Design Details

### Resource index policies

The `ResourceIndexPolicy` resource will be managed by platform operators to
control which resources in the platform are made available to users. The search
service will begin indexing a policies immediately after a policy is created.

The resource index policy will allow users to define the resource applicable to
the index policy and filter resources using CEL expressions. Index policies will
also allow users to configure which fields from the resource are indexed and how
they're indexed (field filtering, full-text search, etc.)

Here's an example index policy that would configure search service to index all
active organizations.

```yaml
apiVersion: search.miloapis.com/v1alpha1
kind: ResourceIndexPolicy
metadata:
  name: organizations
spec:
  targetResource:
    group: resourcemanager.miloapis.com
    version: v1alpha1
    kind: Organization
  conditions:
    # Only index organizations that are active
    - name: ready-only
      expression: status.conditions.exists(c, c.type == 'Ready' && c.status == 'True')
  fields:
    # Index the resource name
    - path: metadata.name
      searchable: true
      filterable: true
    # Index the description of the resource
    - path: metadata.annotations["kubernetes.io/description"]
      searchable: true
      filterable: true
```

The index policy will used a versioned reference to resources since the field
paths for resources may be different between versions. The system should monitor
for deprecated resource versions being referenced in index policies.
