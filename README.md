# Kubernetes API Server Template

> **This is a GitHub template repository.** Click "Use this template" to create a new Kubernetes aggregated API server for your own custom resources.

A production-ready template for building Kubernetes aggregated API servers. This provides all the scaffolding and boilerplate needed to create custom Kubernetes APIs that integrate seamlessly with `kubectl` and the Kubernetes ecosystem.

## What's Included

- **Full API server scaffolding**: Production-ready Kubernetes aggregated API server setup
- **Example custom resource**: `ExampleResource` showing best practices
- **OpenAPI/Swagger integration**: Automatic API documentation generation
- **Metrics & monitoring**: Prometheus metrics built-in
- **Version management**: Git-based version injection
- **Development tooling**: Task-based build system, Docker support
- **Code generation**: Kubernetes code-gen integration for deepcopy and clients

The template includes a working `ExampleResource` as a concrete example. You'll customize this to create your own resources like `DataProcessing`, `BackupJob`, `AnalyticsQuery`, etc.

## Using This Template

### 1. Create Repository
Click "Use this template" on GitHub to create your new repository.

### 2. Find & Replace

Use global find-and-replace (case-sensitive) across all files:

| Find | Replace With | Example |
|------|-------------|---------|
| `github.com/example-org/example-service` | Your module path | `github.com/myorg/myservice` |
| `example.example-org.io` | Your API group | `myresource.mycompany.io` |
| `ghcr.io/example-org/example-service` | Your container registry | `ghcr.io/myorg/myservice` |
| `example-service` | Your service name | `myservice` |
| `ExampleService` | Your service name | `MyService` |
| `ExampleResource` | Your resource type | `DataProcessing` |

### 3. Rename Directories

```bash
mv cmd/example-service cmd/myservice
mv pkg/apis/example-service pkg/apis/myservice
```

### 4. Customize API Types

Edit `pkg/apis/myservice/v1alpha1/types.go`:
- Rename `ExampleResource` to your resource type
- Update `ExampleResourceSpec` fields for your use case
- Update `ExampleResourceStatus` fields for your status
- Review `genclient` directives for your needs (namespaced vs cluster-scoped, allowed verbs)

### 5. Build & Generate

```bash
go mod tidy
task generate
task build
task test
```

### 6. Implement Storage Backend

The main TODO is in `internal/apiserver/apiserver.go` - look for the `TEMPLATE NOTE` comment. You need to:
- Create a REST storage implementation
- Connect to your backend (database, API, cache, etc.)
- Implement CRUD operations
- Register storage in the apiserver

See the [Kubernetes sample-apiserver](https://github.com/kubernetes/sample-apiserver/tree/master/pkg/registry) for examples.

### 7. Deploy (Optional)

Test deployment to a Kubernetes cluster:

```bash
# Review deployment manifests
cat config/README.md

# Generate and review manifests
kubectl kustomize config/base

# Deploy to cluster (requires TLS certs)
./config/deploy-dev.sh
```

See [config/README.md](config/README.md) for detailed deployment instructions.

### 8. Clean Up

Search for `TEMPLATE NOTE` comments throughout the codebase and remove them once you've addressed each customization point.

## Prerequisites

**For users:**
- Kubernetes 1.34+ cluster
- kubectl configured to access your cluster

**For developers:**
- Go 1.25.0 or later
- [Task](https://taskfile.dev) for development workflows
- Docker for building container images

## License

See [LICENSE](LICENSE) for details.

---

**Questions or feedback?** Open an issue—we're here to help!
