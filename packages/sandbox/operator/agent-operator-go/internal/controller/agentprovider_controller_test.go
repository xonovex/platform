package controller

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/test/testutil"
)

func secretBackedProvider() *agentv1alpha1.AgentProvider {
	return testutil.NewAgentProvider("default", "anthropic", testutil.WithAuthTokenSecretRef("api-key", "token"))
}

func apiKeySecret(data map[string][]byte) *corev1.Secret {
	return testutil.NewSecret("default", "api-key", data)
}

func reconcileProvider(t *testing.T, h providerHarness) {
	t.Helper()
	if _, err := h.reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: runKey("anthropic")}); err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
}

func getProvider(t *testing.T, h providerHarness) *agentv1alpha1.AgentProvider {
	t.Helper()
	provider, err := h.store.GetAgentProvider(context.Background(), runKey("anthropic"))
	if err != nil {
		t.Fatalf("get provider: %v", err)
	}
	return provider
}

func TestAgentProviderReconcile_ReportsSecretReadiness(t *testing.T) {
	tests := []struct {
		name     string
		provider *agentv1alpha1.AgentProvider
		objects  []any
		ready    bool
		events   []string
	}{
		{name: "the key resolves", provider: secretBackedProvider(), objects: []any{apiKeySecret(map[string][]byte{"token": []byte("value")})}, ready: true, events: []string{"ProviderSecretResolved"}},
		{name: "the secret is absent", provider: secretBackedProvider(), ready: false, events: []string{"ProviderSecretMissing"}},
		{name: "the key is absent", provider: secretBackedProvider(), objects: []any{apiKeySecret(map[string][]byte{"other": []byte("value")})}, ready: false, events: []string{"ProviderSecretMissing"}},
		{name: "no secret is referenced", provider: testutil.NewAgentProvider("default", "anthropic", testutil.WithAuthTokenEnv("ANTHROPIC_API_KEY")), ready: true, events: []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newProviderHarness(append([]any{test.provider}, test.objects...)...)

			reconcileProvider(t, h)

			provider := getProvider(t, h)
			if provider.Status.Ready != test.ready {
				t.Errorf("Ready = %v, want %v", provider.Status.Ready, test.ready)
			}
			condition := meta.FindStatusCondition(provider.Status.Conditions, "Ready")
			if condition == nil || !condition.LastTransitionTime.Time.Equal(testEpoch) {
				t.Errorf("Ready condition = %+v, want one stamped with the clock time", condition)
			}
			if !slices.Equal(h.events.reasons(), test.events) {
				t.Errorf("events = %v, want %v", h.events.reasons(), test.events)
			}
		})
	}
}

// Only transitions are reported and written, so a resync over unchanged state
// stays quiet and keeps the transition time.
func TestAgentProviderReconcile_LeavesAStableStatusUnchanged(t *testing.T) {
	h := newProviderHarness(secretBackedProvider(), apiKeySecret(map[string][]byte{"token": []byte("value")}))
	reconcileProvider(t, h)
	h.clock.advance(time.Hour)

	reconcileProvider(t, h)

	if h.store.StatusWrites() != 1 {
		t.Errorf("status writes = %d, want one", h.store.StatusWrites())
	}
	if len(h.events.events) != 1 {
		t.Errorf("events = %v, want only the first transition", h.events.reasons())
	}
	condition := meta.FindStatusCondition(getProvider(t, h).Status.Conditions, "Ready")
	if !condition.LastTransitionTime.Time.Equal(testEpoch) {
		t.Errorf("transition time = %v, want the first reconcile time", condition.LastTransitionTime)
	}
}

func TestAgentProviderReconcile_IgnoresAnAbsentProvider(t *testing.T) {
	h := newProviderHarness()

	reconcileProvider(t, h)

	if h.store.StatusWrites() != 0 {
		t.Error("an absent provider was written")
	}
}

func TestAgentProviderReconcile_SkipsAProviderBeingDeleted(t *testing.T) {
	provider := secretBackedProvider()
	deletionTimestamp := metav1.NewTime(testEpoch)
	provider.DeletionTimestamp = &deletionTimestamp
	h := newProviderHarness(provider)

	reconcileProvider(t, h)

	if len(getProvider(t, h).Status.Conditions) != 0 {
		t.Error("a provider being deleted gained conditions")
	}
}

func TestAgentProviderReconcile_SurfacesStoreFailures(t *testing.T) {
	for _, method := range []string{"GetAgentProvider", "GetSecret", "UpdateAgentProviderStatus"} {
		t.Run(method, func(t *testing.T) {
			h := newProviderHarness(secretBackedProvider())
			h.store.Fail(method, errInjected)

			_, err := h.reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: runKey("anthropic")})

			if !errors.Is(err, errInjected) {
				t.Errorf("Reconcile() error = %v, want the injected failure", err)
			}
			if len(h.events.events) != 0 {
				t.Errorf("events = %v, want none before a successful write", h.events.reasons())
			}
		})
	}
}

func TestProvidersForSecret_EnqueuesOnlyMatchingProviders(t *testing.T) {
	h := newProviderHarness(
		testutil.NewAgentProvider("default", "matching", testutil.WithAuthTokenSecretRef("api-key", "token")),
		testutil.NewAgentProvider("default", "other-secret", testutil.WithAuthTokenSecretRef("different-key", "token")),
		testutil.NewAgentProvider("default", "no-secret", testutil.WithAuthTokenEnv("ANTHROPIC_API_KEY")),
		testutil.NewAgentProvider("other", "elsewhere", testutil.WithAuthTokenSecretRef("api-key", "token")),
	)

	requests := h.reconciler.providersForSecret(context.Background(), apiKeySecret(nil))

	want := []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: "default", Name: "matching"}}}
	if !slices.Equal(requests, want) {
		t.Errorf("providersForSecret() = %v, want %v", requests, want)
	}
}

func TestProvidersForSecret_IgnoresANonSecretObject(t *testing.T) {
	h := newProviderHarness(secretBackedProvider())

	requests := h.reconciler.providersForSecret(context.Background(),
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "api-key", Namespace: "default"}})

	if requests != nil {
		t.Errorf("providersForSecret() = %v, want nil for a non-Secret object", requests)
	}
}

// A failed list enqueues nothing; the next Secret event or resync retries.
func TestProvidersForSecret_EnqueuesNothingWhenTheListFails(t *testing.T) {
	h := newProviderHarness(secretBackedProvider())
	h.store.Fail("ListAgentProviders", errInjected)

	if requests := h.reconciler.providersForSecret(context.Background(), apiKeySecret(nil)); requests != nil {
		t.Errorf("providersForSecret() = %v, want nil", requests)
	}
}
