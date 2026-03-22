package builder

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestDefaultContainerSecurityContext_NilOverride(t *testing.T) {
	sc := DefaultContainerSecurityContext(nil)

	if sc.AllowPrivilegeEscalation == nil || *sc.AllowPrivilegeEscalation != false {
		t.Error("AllowPrivilegeEscalation should be false")
	}
	if sc.RunAsNonRoot == nil || *sc.RunAsNonRoot != true {
		t.Error("RunAsNonRoot should be true")
	}
	if sc.ReadOnlyRootFilesystem == nil || *sc.ReadOnlyRootFilesystem != true {
		t.Error("ReadOnlyRootFilesystem should be true")
	}
	if sc.SeccompProfile == nil || sc.SeccompProfile.Type != corev1.SeccompProfileTypeRuntimeDefault {
		t.Error("SeccompProfile should be RuntimeDefault")
	}
	if sc.Capabilities == nil || len(sc.Capabilities.Drop) != 1 || sc.Capabilities.Drop[0] != "ALL" {
		t.Error("Capabilities.Drop should be [ALL]")
	}
}

func TestDefaultContainerSecurityContext_WithOverride(t *testing.T) {
	allowPrivEsc := true
	uid := int64(1000)
	gid := int64(1000)
	override := &corev1.SecurityContext{
		AllowPrivilegeEscalation: &allowPrivEsc,
		RunAsUser:                &uid,
		RunAsGroup:               &gid,
		Capabilities:             &corev1.Capabilities{Add: []corev1.Capability{"NET_BIND_SERVICE"}},
	}

	sc := DefaultContainerSecurityContext(override)

	if *sc.AllowPrivilegeEscalation != true {
		t.Error("AllowPrivilegeEscalation override should be true")
	}
	if sc.RunAsUser == nil || *sc.RunAsUser != 1000 {
		t.Error("RunAsUser should be 1000")
	}
	if sc.RunAsGroup == nil || *sc.RunAsGroup != 1000 {
		t.Error("RunAsGroup should be 1000")
	}
	if sc.Capabilities.Add[0] != "NET_BIND_SERVICE" {
		t.Error("Capabilities should be overridden")
	}
	// Defaults should still be present for non-overridden fields
	if *sc.RunAsNonRoot != true {
		t.Error("RunAsNonRoot default should be preserved")
	}
	if *sc.ReadOnlyRootFilesystem != true {
		t.Error("ReadOnlyRootFilesystem default should be preserved")
	}
}

func TestDefaultPodSecurityContext_NilOverride(t *testing.T) {
	psc := DefaultPodSecurityContext(nil)

	if psc.RunAsNonRoot == nil || *psc.RunAsNonRoot != true {
		t.Error("RunAsNonRoot should be true")
	}
	if psc.SeccompProfile == nil || psc.SeccompProfile.Type != corev1.SeccompProfileTypeRuntimeDefault {
		t.Error("SeccompProfile should be RuntimeDefault")
	}
}

func TestDefaultPodSecurityContext_WithOverride(t *testing.T) {
	uid := int64(2000)
	gid := int64(2000)
	fsGroup := int64(3000)
	override := &corev1.PodSecurityContext{
		RunAsUser:  &uid,
		RunAsGroup: &gid,
		FSGroup:    &fsGroup,
	}

	psc := DefaultPodSecurityContext(override)

	if psc.RunAsUser == nil || *psc.RunAsUser != 2000 {
		t.Error("RunAsUser should be 2000")
	}
	if psc.RunAsGroup == nil || *psc.RunAsGroup != 2000 {
		t.Error("RunAsGroup should be 2000")
	}
	if psc.FSGroup == nil || *psc.FSGroup != 3000 {
		t.Error("FSGroup should be 3000")
	}
	// Defaults preserved
	if *psc.RunAsNonRoot != true {
		t.Error("RunAsNonRoot default should be preserved")
	}
}
