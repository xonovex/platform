package builder

import corev1 "k8s.io/api/core/v1"

// DefaultContainerSecurityContext returns hardened defaults matching the operator manager pod.
// Fields from override take precedence over defaults.
func DefaultContainerSecurityContext(override *corev1.SecurityContext) *corev1.SecurityContext {
	allowPrivEsc := false
	runAsNonRoot := true
	readOnlyRoot := true
	seccomp := corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}
	dropAll := []corev1.Capability{"ALL"}

	sc := &corev1.SecurityContext{
		AllowPrivilegeEscalation: &allowPrivEsc,
		RunAsNonRoot:             &runAsNonRoot,
		ReadOnlyRootFilesystem:   &readOnlyRoot,
		SeccompProfile:           &seccomp,
		Capabilities:             &corev1.Capabilities{Drop: dropAll},
	}

	if override != nil {
		if override.AllowPrivilegeEscalation != nil {
			sc.AllowPrivilegeEscalation = override.AllowPrivilegeEscalation
		}
		if override.RunAsNonRoot != nil {
			sc.RunAsNonRoot = override.RunAsNonRoot
		}
		if override.ReadOnlyRootFilesystem != nil {
			sc.ReadOnlyRootFilesystem = override.ReadOnlyRootFilesystem
		}
		if override.SeccompProfile != nil {
			sc.SeccompProfile = override.SeccompProfile
		}
		if override.Capabilities != nil {
			sc.Capabilities = override.Capabilities
		}
		if override.RunAsUser != nil {
			sc.RunAsUser = override.RunAsUser
		}
		if override.RunAsGroup != nil {
			sc.RunAsGroup = override.RunAsGroup
		}
	}

	return sc
}

// DefaultPodSecurityContext returns hardened pod-level defaults.
// Fields from override take precedence over defaults.
func DefaultPodSecurityContext(override *corev1.PodSecurityContext) *corev1.PodSecurityContext {
	runAsNonRoot := true
	seccomp := corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}

	psc := &corev1.PodSecurityContext{
		RunAsNonRoot:   &runAsNonRoot,
		SeccompProfile: &seccomp,
	}

	if override != nil {
		if override.RunAsNonRoot != nil {
			psc.RunAsNonRoot = override.RunAsNonRoot
		}
		if override.RunAsUser != nil {
			psc.RunAsUser = override.RunAsUser
		}
		if override.RunAsGroup != nil {
			psc.RunAsGroup = override.RunAsGroup
		}
		if override.FSGroup != nil {
			psc.FSGroup = override.FSGroup
		}
		if override.SeccompProfile != nil {
			psc.SeccompProfile = override.SeccompProfile
		}
		if override.Sysctls != nil {
			psc.Sysctls = override.Sysctls
		}
	}

	return psc
}
