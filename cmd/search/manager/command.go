package manager

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.miloapis.net/search/pkg/apis/policy/install"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	policycontroller "go.miloapis.net/search/internal/controllers/policy"
	policyv1alpha1webhook "go.miloapis.net/search/internal/webhooks/policy/v1alpha1"
	policyv1alpha1 "go.miloapis.net/search/pkg/apis/policy/v1alpha1"

	"go.miloapis.net/search/internal/cel"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	install.Install(scheme)
}

// ControllerManagerOptions contains configuration for the controller manager
type ControllerManagerOptions struct {
	MetricsAddr          string
	EnableLeaderElection bool
	ProbeAddr            string
	SecureMetrics        bool
	EnableHTTP2          bool
	WebhookCertPath      string
	WebhookCertName      string
	WebhookCertKey       string
	MaxCELDepth          int
}

// NewControllerManagerOptions creates a new ControllerManagerOptions with default values
func NewControllerManagerOptions() *ControllerManagerOptions {
	return &ControllerManagerOptions{
		MetricsAddr:          ":8080",
		ProbeAddr:            ":8081",
		EnableLeaderElection: false,
		SecureMetrics:        false,
		EnableHTTP2:          false,
		WebhookCertName:      "tls.crt",
		WebhookCertKey:       "tls.key",
		MaxCELDepth:          50,
	}
}

// AddFlags adds flags to the specified FlagSet
func (o *ControllerManagerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.MetricsAddr, "metrics-bind-address", o.MetricsAddr, "The address the metric endpoint binds to.")
	fs.StringVar(&o.ProbeAddr, "health-probe-bind-address", o.ProbeAddr, "The address the probe endpoint binds to.")
	fs.BoolVar(&o.EnableLeaderElection, "leader-elect", o.EnableLeaderElection,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.BoolVar(&o.SecureMetrics, "metrics-secure", o.SecureMetrics,
		"If set the metrics endpoint is served securely")
	fs.BoolVar(&o.EnableHTTP2, "enable-http2", o.EnableHTTP2,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	fs.StringVar(&o.WebhookCertPath, "webhook-cert-path", o.WebhookCertPath, "The directory that contains the webhook certificate.")
	fs.StringVar(&o.WebhookCertName, "webhook-cert-name", o.WebhookCertName, "The name of the webhook certificate file.")
	fs.StringVar(&o.WebhookCertKey, "webhook-cert-key", o.WebhookCertKey, "The name of the webhook key file.")
	fs.IntVar(&o.MaxCELDepth, "max-cel-depth", o.MaxCELDepth, "Maximum recursion depth allowed for CEL expressions in policies.")
}

// Validate validates the options
func (o *ControllerManagerOptions) Validate() error {
	if len(o.WebhookCertPath) == 0 {
		return fmt.Errorf("missing required flag: --webhook-cert-path")
	}
	if o.MaxCELDepth < 1 {
		return fmt.Errorf("max-cel-depth must be greater than 0")
	}
	return nil
}

// Complete completes the options
func (o *ControllerManagerOptions) Complete() error {
	return nil
}

// NewControllerManagerCommand creates the controller-manager subcommand.
func NewControllerManagerCommand() *cobra.Command {
	o := NewControllerManagerOptions()

	cmd := &cobra.Command{
		Use:   "controller-manager",
		Short: "Start the controller manager",
		Long:  `Start the controller manager to reconcile and validate resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			return Run(o, cmd.Context())
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}

// Run starts the controller manager
func Run(o *ControllerManagerOptions, ctx context.Context) error {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	var tlsOpts []func(*tls.Config)
	if !o.EnableHTTP2 {
		tlsOpts = append(tlsOpts, func(c *tls.Config) {
			c.NextProtos = []string{"http/1.1"}
		})
	}

	// Webhook TLS configuration
	var webhookCertWatcher *certwatcher.CertWatcher
	webhookTLSOpts := tlsOpts

	if len(o.WebhookCertPath) > 0 {
		setupLog.Info("Initializing webhook certificate watcher",
			"path", o.WebhookCertPath, "name", o.WebhookCertName, "key", o.WebhookCertKey)

		var err error
		webhookCertWatcher, err = certwatcher.New(
			filepath.Join(o.WebhookCertPath, o.WebhookCertName),
			filepath.Join(o.WebhookCertPath, o.WebhookCertKey),
		)
		if err != nil {
			setupLog.Error(err, "Failed to initialize webhook certificate watcher")
			os.Exit(1)
		}

		webhookTLSOpts = append(webhookTLSOpts, func(config *tls.Config) {
			config.GetCertificate = webhookCertWatcher.GetCertificate
		})
	} else {
		// If no cert path provided, ensure empty CertDir to trigger controller-runtime's
		// default behavior (which might be self-signed or just failing if certs needed but not found).
		// However, NewServer logic is: if CertDir is empty, it defaults to /tmp/k8s-webhook-server/serving-certs.
		// To force it to generate self-signed certs in a temp dir if not found, we rely on default behavior
		// but we explicitly log that we are using self-signed/default certs.
		setupLog.Info("No webhook cert path provided, using default self-signed certs or existing certs in /tmp/k8s-webhook-server/serving-certs")
	}

	webhookServerOptions := webhook.Options{
		Port:    9443,
		TLSOpts: webhookTLSOpts,
	}

	webhookServer := webhook.NewServer(webhookServerOptions)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: o.MetricsAddr, SecureServing: o.SecureMetrics, TLSOpts: tlsOpts},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: o.ProbeAddr,
		LeaderElection:         o.EnableLeaderElection,
		LeaderElectionID:       "1x3u92ac.datumapis.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if webhookCertWatcher != nil {
		setupLog.Info("Adding webhook certificate watcher to manager")
		if err := mgr.Add(webhookCertWatcher); err != nil {
			setupLog.Error(err, "unable to add webhook certificate watcher to manager")
			os.Exit(1)
		}
	}

	// Register Webhook
	celValidator, err := cel.NewValidator(o.MaxCELDepth)
	if err != nil {
		setupLog.Error(err, "unable to create CEL validator")
		os.Exit(1)
	}

	if err = (&policycontroller.ResourceIndexPolicyReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		CelValidator: celValidator,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ResourceIndexPolicy")
		os.Exit(1)
	}

	if err = ctrl.NewWebhookManagedBy(mgr, &policyv1alpha1.ResourceIndexPolicy{}).
		WithValidator(&policyv1alpha1webhook.ResourceIndexPolicyValidator{
			CelValidator: celValidator,
		}).
		Complete(); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ResourceIndexPolicy")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
	return nil
}
