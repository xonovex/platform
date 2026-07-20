package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/controller"
	runtel "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/telemetry"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = agentv1alpha1.AddToScheme(scheme)
}

func main() {
	var probeAddr string
	var enableLeaderElection bool
	var decisionServiceURL string
	var triggerBindAddress string
	var remediationTriggerURL string
	var remediationTokenFile string
	var escalationRouterURL string
	var escalationRouterTokenFile string
	var escalationResponseBaseURL string
	var workspaceInitImage string

	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	flag.StringVar(&decisionServiceURL, "decision-service-url", decisionServiceURLFromEnvironment(), "Governance decision service base URL.")
	flag.StringVar(&triggerBindAddress, "trigger-bind-address", ":8090", "The address the authenticated AgentTrigger receiver binds to.")
	flag.StringVar(&remediationTriggerURL, "remediation-trigger-url", "", "AgentTrigger endpoint used for drift remediation runs.")
	flag.StringVar(&remediationTokenFile, "remediation-trigger-token-file", "", "File containing the AgentTrigger bearer token for drift remediation.")
	flag.StringVar(&escalationRouterURL, "escalation-router-url", "", "Provider-native endpoint that delivers accountable-recipient escalation requests.")
	flag.StringVar(&escalationRouterTokenFile, "escalation-router-token-file", "", "File containing the bearer token for the escalation router.")
	flag.StringVar(&escalationResponseBaseURL, "escalation-response-base-url", "", "Externally reachable base URL for authenticated escalation responses.")
	flag.StringVar(&workspaceInitImage, "workspace-init-image", controller.DefaultWorkspaceInitImage, "Digest-pinned image used to initialize shared workspaces.")

	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	if err := validator.ValidatePinnedImageReference(workspaceInitImage); err != nil {
		setupLog.Error(err, "invalid workspace init image", "image", workspaceInitImage)
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "agent-operator.xonovex.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	remediationRouter, err := createRemediationRouter(remediationTriggerURL, remediationTokenFile)
	if err != nil {
		setupLog.Error(err, "unable to configure remediation trigger")
		os.Exit(1)
	}
	escalationRouter, err := createEscalationRouter(escalationRouterURL, escalationRouterTokenFile, escalationResponseBaseURL)
	if err != nil {
		setupLog.Error(err, "unable to configure escalation router")
		os.Exit(1)
	}
	if err = (&controller.AgentRunReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		Recorder:          mgr.GetEventRecorder("agent-operator"),
		Telemetry:         runtel.NewOTelSink(),
		RemediationRouter: remediationRouter,
		EscalationRouter:  escalationRouter,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AgentRun")
		os.Exit(1)
	}

	if err = (&controller.AgentProviderReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorder("agent-operator"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AgentProvider")
		os.Exit(1)
	}

	if err = (&controller.AgentWorkspaceReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		Recorder:           mgr.GetEventRecorder("agent-operator"),
		WorkspaceInitImage: workspaceInitImage,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AgentWorkspace")
		os.Exit(1)
	}

	if err = (&controller.AgentScheduleReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AgentSchedule")
		os.Exit(1)
	}

	if err = (&controller.AgentTriggerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AgentTrigger")
		os.Exit(1)
	}

	if err = mgr.Add(&controller.AgentTriggerReceiver{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Address: triggerBindAddress,
	}); err != nil {
		setupLog.Error(err, "unable to add AgentTrigger receiver")
		os.Exit(1)
	}

	if err = (&webhook.AgentRunWebhook{
		Client:         mgr.GetClient(),
		DecisionClient: webhook.NewHTTPGovernanceDecisionClient(decisionServiceURL),
	}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to set up webhook", "webhook", "AgentRun")
		os.Exit(1)
	}
	if err = (&webhook.AgentHarnessWebhook{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to set up webhook", "webhook", "AgentHarness")
		os.Exit(1)
	}
	if err = (&webhook.AgentProviderWebhook{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to set up webhook", "webhook", "AgentProvider")
		os.Exit(1)
	}
	if err = (&webhook.AgentToolchainWebhook{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to set up webhook", "webhook", "AgentToolchain")
		os.Exit(1)
	}
	if err = (&webhook.AgentWorkspaceWebhook{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to set up webhook", "webhook", "AgentWorkspace")
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
}

func createRemediationRouter(endpoint, tokenFile string) (controller.RemediationRouter, error) {
	if endpoint == "" && tokenFile == "" {
		return nil, nil
	}
	if endpoint == "" || tokenFile == "" {
		return nil, fmt.Errorf("remediation trigger URL and token file must be configured together")
	}
	token, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("read remediation trigger token file: %w", err)
	}
	if strings.TrimSpace(string(token)) == "" {
		return nil, fmt.Errorf("remediation trigger token file is empty")
	}
	return controller.NewAgentTriggerRemediationRouter(endpoint, strings.TrimSpace(string(token))), nil
}

func createEscalationRouter(endpoint, tokenFile, responseBaseURL string) (controller.EscalationRouter, error) {
	if endpoint == "" && tokenFile == "" && responseBaseURL == "" {
		return nil, nil
	}
	if endpoint == "" || tokenFile == "" || responseBaseURL == "" {
		return nil, fmt.Errorf("escalation router URL, token file, and response base URL must be configured together")
	}
	token, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("read escalation router token file: %w", err)
	}
	if strings.TrimSpace(string(token)) == "" {
		return nil, fmt.Errorf("escalation router token file is empty")
	}
	return controller.NewHTTPEscalationRouter(endpoint, strings.TrimSpace(string(token)), responseBaseURL), nil
}

func decisionServiceURLFromEnvironment() string {
	if value := os.Getenv("DECISION_SERVICE_URL"); value != "" {
		return value
	}
	return "http://127.0.0.1:8787"
}
