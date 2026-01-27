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
  # Identifies the resource type this policy applies to. Uses a versioned
  # reference since field paths may differ between API versions.
  targetResource:
    group: resourcemanager.miloapis.com
    version: v1alpha1
    kind: Organization

  # CEL expressions that filter which resources are indexed. Multiple
  # conditions can be specified and are evaluated with OR semantics - a
  # resource is indexed if it satisfies ANY condition. Use && within a
  # single expression to require multiple criteria together.
  #
  # Each condition has:
  # - name: A unique identifier for the condition, used in status reporting
  #   and debugging to identify which condition(s) matched a resource.
  # - expression: A CEL expression that must evaluate to a boolean. The
  #   resource is available as the root object in the expression context.
  #
  # Available CEL operations:
  # - Field access: spec.replicas, metadata.name, status.phase
  # - Map access: metadata.labels["app"], metadata.annotations["key"]
  # - Comparisons: ==, !=, <, <=, >, >=
  # - Logical operators: &&, ||, !
  # - String functions: contains(), startsWith(), endsWith(), matches()
  # - List functions: exists(), all(), size(), map(), filter()
  # - Membership: "value" in list, "key" in map
  # - Ternary: condition ? trueValue : falseValue
  conditions:
    # Index resources that are ready and not being deleted
    - name: active-resources
      expression: |
        status.conditions.exists(c, c.type == 'Ready' && c.status == 'True')
        && !has(metadata.deletionTimestamp)
    # Also index resources in production namespaces regardless of status
    - name: production-resources
      expression: metadata.namespace.startsWith("prod-")

  # Defines which fields from the resource are indexed and how they behave
  # in search operations.
  fields:
    # The JSONPath to the field value in the resource. Supports nested paths
    # and map key access using bracket notation.
    - path: metadata.name
      # When true, the field value is included in full-text search operations.
      # The value is tokenized and analyzed for relevance-based matching.
      searchable: true
      # When true, the field can be used in filter expressions for exact
      # matching, range queries, and other structured filtering operations.
      filterable: true
      # When true, the search service will return aggregated counts of unique
      # values for this field. Enables clients to discover available filter
      # values and build faceted navigation interfaces.
      facetable: true

    - path: metadata.annotations["kubernetes.io/description"]
      searchable: true
      filterable: true
      facetable: false
```

The index policy will used a versioned reference to resources since the field
paths for resources may be different between versions. The system should monitor
for deprecated resource versions being referenced in index policies.

#### Condition evaluation

Conditions provide fine-grained control over which resource instances are
indexed. When multiple conditions are specified, they are evaluated using OR
semantics - a resource is indexed if it satisfies ANY condition. This allows
defining multiple independent criteria for inclusion.

Use `&&` within a single CEL expression when you need AND semantics:

```yaml
conditions:
  - name: ready-in-production
    expression: |
      status.conditions.exists(c, c.type == 'Ready' && c.status == 'True')
      && metadata.namespace.startsWith("prod-")
```

Conditions are re-evaluated when resources change. If a resource no longer
satisfies any condition (e.g., it transitions from Ready to NotReady), it will
be removed from the search index. Similarly, resources that begin satisfying
a condition after an update will be added to the index.
