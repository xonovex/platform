package controller

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func newProviderReconciler(
	recorder events.EventRecorder,
	objects ...client.Object,
) (*AgentProviderReconciler, client.Client) {
	scheme := testutil.NewScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentProvider{}).
		WithObjects(objects...).
		Build()
	return &AgentProviderReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Recorder: recorder,
	}, fakeClient
}

func reconcileProvider(
	t *testing.T,
	reconciler *AgentProviderReconciler,
	namespace, name string,
) {
	t.Helper()
	if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
	}); err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
}

func providerReadiness(
	t *testing.T,
	c client.Client,
	namespace, name string,
) agentv1alpha1.AgentProvider {
	t.Helper()
	var provider agentv1alpha1.AgentProvider
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: namespace, Name: name,
	}, &provider); err != nil {
		t.Fatalf("Get provider: %v", err)
	}
	return provider
}

func TestAgentProviderReconcile_MarksReadyWhenSecretKeyResolves(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenSecretRef("api-key", "token"))
	secret := testutil.NewSecret("default", "api-key", map[string][]byte{
		"token": []byte("secret-value"),
	})
	reconciler, c := newProviderReconciler(events.NewFakeRecorder(10), provider, secret)

	reconcileProvider(t, reconciler, "default", "anthropic")

	resolved := providerReadiness(t, c, "default", "anthropic")
	if !resolved.Status.Ready {
		t.Error("provider must be ready when the secret key resolves")
	}
}

func TestAgentProviderReconcile_MarksNotReadyWhenSecretIsAbsent(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenSecretRef("api-key", "token"))
	reconciler, c := newProviderReconciler(events.NewFakeRecorder(10), provider)

	reconcileProvider(t, reconciler, "default", "anthropic")

	resolved := providerReadiness(t, c, "default", "anthropic")
	if resolved.Status.Ready {
		t.Error("provider must not be ready when the referenced secret is absent")
	}
}

// The secret existing is not enough: the named key has to be in it.
func TestAgentProviderReconcile_MarksNotReadyWhenSecretLacksTheKey(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenSecretRef("api-key", "token"))
	secret := testutil.NewSecret("default", "api-key", map[string][]byte{
		"other": []byte("secret-value"),
	})
	reconciler, c := newProviderReconciler(events.NewFakeRecorder(10), provider, secret)

	reconcileProvider(t, reconciler, "default", "anthropic")

	resolved := providerReadiness(t, c, "default", "anthropic")
	if resolved.Status.Ready {
		t.Error("provider must not be ready when the secret lacks the referenced key")
	}
}

// A provider authenticating by environment variable references no secret, so
// there is nothing to resolve and it is ready outright.
func TestAgentProviderReconcile_MarksReadyWithoutASecretRef(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenEnv("ANTHROPIC_API_KEY"))
	reconciler, c := newProviderReconciler(events.NewFakeRecorder(10), provider)

	reconcileProvider(t, reconciler, "default", "anthropic")

	resolved := providerReadiness(t, c, "default", "anthropic")
	if !resolved.Status.Ready {
		t.Error("a provider with no secret reference must be ready")
	}
}

func TestAgentProviderReconcile_IgnoresAnAbsentProvider(t *testing.T) {
	reconciler, _ := newProviderReconciler(events.NewFakeRecorder(10))

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "absent"},
	})

	if err != nil {
		t.Fatalf("Reconcile() error = %v, want nil for a deleted provider", err)
	}
	if result.RequeueAfter != 0 {
		t.Error("Reconcile() must not requeue for a deleted provider")
	}
}

func TestAgentProviderReconcile_SkipsAProviderBeingDeleted(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenSecretRef("api-key", "token"))
	deletionTimestamp := metav1.Now()
	provider.DeletionTimestamp = &deletionTimestamp
	provider.Finalizers = []string{"xonovex.com/test"}
	reconciler, c := newProviderReconciler(events.NewFakeRecorder(10), provider)

	reconcileProvider(t, reconciler, "default", "anthropic")

	resolved := providerReadiness(t, c, "default", "anthropic")
	if len(resolved.Status.Conditions) > 0 {
		t.Errorf("a provider being deleted must not gain conditions, got %v",
			resolved.Status.Conditions)
	}
}

func TestAgentProviderReconcile_EmitsSecretResolvedOnTransition(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenSecretRef("api-key", "token"))
	secret := testutil.NewSecret("default", "api-key", map[string][]byte{
		"token": []byte("secret-value"),
	})
	recorder := events.NewFakeRecorder(10)
	reconciler, _ := newProviderReconciler(recorder, provider, secret)

	reconcileProvider(t, reconciler, "default", "anthropic")

	select {
	case event := <-recorder.Events:
		if !strings.Contains(event, "ProviderSecretResolved") {
			t.Errorf("event = %q, want a ProviderSecretResolved event", event)
		}
	default:
		t.Fatal("expected an event for the readiness transition")
	}
}

func TestAgentProviderReconcile_EmitsSecretMissingOnTransition(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenSecretRef("api-key", "token"))
	recorder := events.NewFakeRecorder(10)
	reconciler, _ := newProviderReconciler(recorder, provider)

	reconcileProvider(t, reconciler, "default", "anthropic")

	select {
	case event := <-recorder.Events:
		if !strings.Contains(event, "ProviderSecretMissing") {
			t.Errorf("event = %q, want a ProviderSecretMissing event", event)
		}
	default:
		t.Fatal("expected an event for the readiness transition")
	}
}

// Only transitions are reported, so a second reconcile over unchanged state
// stays quiet rather than repeating the event every resync.
func TestAgentProviderReconcile_EmitsNoEventWithoutATransition(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenSecretRef("api-key", "token"))
	secret := testutil.NewSecret("default", "api-key", map[string][]byte{
		"token": []byte("secret-value"),
	})
	recorder := events.NewFakeRecorder(10)
	reconciler, _ := newProviderReconciler(recorder, provider, secret)

	reconcileProvider(t, reconciler, "default", "anthropic")
	<-recorder.Events
	reconcileProvider(t, reconciler, "default", "anthropic")

	select {
	case event := <-recorder.Events:
		t.Errorf("unchanged readiness emitted %q, want no event", event)
	default:
	}
}

func TestRecordSecretTransition_StaysSilentWithoutARecorder(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenSecretRef("api-key", "token"))
	reconciler := &AgentProviderReconciler{Scheme: testutil.NewScheme()}

	// The absence of a panic is the assertion: a nil recorder is tolerated.
	reconciler.recordSecretTransition(provider, true)
}

func TestRecordSecretTransition_StaysSilentWithoutASecretRef(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenEnv("ANTHROPIC_API_KEY"))
	recorder := events.NewFakeRecorder(10)
	reconciler := &AgentProviderReconciler{Scheme: testutil.NewScheme(), Recorder: recorder}

	reconciler.recordSecretTransition(provider, false)

	select {
	case event := <-recorder.Events:
		t.Errorf("a provider with no secret ref emitted %q, want no event", event)
	default:
	}
}

func TestProvidersForSecret_EnqueuesOnlyMatchingProviders(t *testing.T) {
	matching := testutil.NewAgentProvider("default", "matching",
		testutil.WithAuthTokenSecretRef("api-key", "token"))
	otherSecret := testutil.NewAgentProvider("default", "other-secret",
		testutil.WithAuthTokenSecretRef("different-key", "token"))
	noSecret := testutil.NewAgentProvider("default", "no-secret",
		testutil.WithAuthTokenEnv("ANTHROPIC_API_KEY"))
	reconciler, _ := newProviderReconciler(
		events.NewFakeRecorder(10), matching, otherSecret, noSecret,
	)
	secret := testutil.NewSecret("default", "api-key", map[string][]byte{
		"token": []byte("secret-value"),
	})

	requests := reconciler.providersForSecret(context.Background(), secret)

	want := []reconcile.Request{{NamespacedName: types.NamespacedName{
		Namespace: "default", Name: "matching",
	}}}
	if len(requests) != len(want) || requests[0] != want[0] {
		t.Errorf("providersForSecret() = %v, want %v", requests, want)
	}
}

// The watch is namespace scoped, so a same-named secret elsewhere must not wake
// this namespace's providers.
func TestProvidersForSecret_IgnoresOtherNamespaces(t *testing.T) {
	provider := testutil.NewAgentProvider("default", "anthropic",
		testutil.WithAuthTokenSecretRef("api-key", "token"))
	reconciler, _ := newProviderReconciler(events.NewFakeRecorder(10), provider)
	secret := testutil.NewSecret("other", "api-key", map[string][]byte{
		"token": []byte("secret-value"),
	})

	requests := reconciler.providersForSecret(context.Background(), secret)

	if len(requests) != 0 {
		t.Errorf("providersForSecret() = %v, want none across namespaces", requests)
	}
}

func TestProvidersForSecret_IgnoresANonSecretObject(t *testing.T) {
	reconciler, _ := newProviderReconciler(events.NewFakeRecorder(10))

	requests := reconciler.providersForSecret(
		context.Background(),
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
			Name: "api-key", Namespace: "default",
		}},
	)

	if requests != nil {
		t.Errorf("providersForSecret() = %v, want nil for a non-Secret object", requests)
	}
}

func TestUpdateProviderStatus_IgnoresADeletedProvider(t *testing.T) {
	reconciler, _ := newProviderReconciler(events.NewFakeRecorder(10))
	absent := testutil.NewAgentProvider("default", "absent")
	absent.Status.Ready = true

	if err := reconciler.updateProviderStatus(context.Background(), absent); err != nil {
		t.Errorf("updateProviderStatus() error = %v, want nil for a deleted provider", err)
	}
}
