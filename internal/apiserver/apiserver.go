package apiserver

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/klog/v2"

	_ "github.com/example-org/example-service/internal/metrics"
	"github.com/example-org/example-service/pkg/apis/example-service/install"
	"github.com/example-org/example-service/pkg/apis/example-service/v1alpha1"
)

var (
	// Scheme defines the runtime type system for API object serialization.
	Scheme = runtime.NewScheme()
	// Codecs provides serializers for API objects.
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	install.Install(Scheme)

	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// Register unversioned meta types required by the API machinery.
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

// ExtraConfig extends the generic apiserver configuration with search-specific settings.
type ExtraConfig struct {
	// Add custom configuration here as needed
}

// Config combines generic and search-specific configuration.
type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

// ExampleServiceServer is the search aggregated apiserver.
type ExampleServiceServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

// CompletedConfig prevents incomplete configuration from being used.
// Embeds a private pointer that can only be created via Complete().
type CompletedConfig struct {
	*completedConfig
}

// Complete validates and fills default values for the configuration.
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		&cfg.ExtraConfig,
	}

	return CompletedConfig{&c}
}

// New creates and initializes the ExampleServiceServer with storage and API groups.
func (c completedConfig) New() (*ExampleServiceServer, error) {
	genericServer, err := c.GenericConfig.New("search-apiserver", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	s := &ExampleServiceServer{
		GenericAPIServer: genericServer,
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(v1alpha1.GroupName, Scheme, metav1.ParameterCodec, Codecs)

	v1alpha1Storage := map[string]rest.Storage{}

	// TEMPLATE NOTE: This is where you register your REST storage implementations.
	// You need to create a storage backend that implements rest.Storage interface.
	//
	// Example pattern:
	//   storage := registry.NewYourResourceStorage(/* backend config */)
	//   v1alpha1Storage["yourresources"] = storage
	//
	// The storage implementation should handle:
	// - Connecting to your backend (database, API, cache, etc.)
	// - CRUD operations (Create, Get, List, Update, Delete, Watch)
	// - Validation and business logic
	//
	// See kubernetes sample-apiserver for reference implementations:
	// https://github.com/kubernetes/sample-apiserver/tree/master/pkg/registry
	//
	// TODO: Add your storage implementations here

	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1Storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	klog.Info("ExampleService server initialized successfully")

	return s, nil
}

// Run starts the server.
func (s *ExampleServiceServer) Run(ctx context.Context) error {
	return s.GenericAPIServer.PrepareRun().RunWithContext(ctx)
}
