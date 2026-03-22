package builder

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

const (
	nixVolumeName     = "nix-env"
	nixMountPath      = "/nix"
	nixProfileBinPath = "/nix/var/nix/profiles/agent/bin"
	defaultNixImage   = "nixos/nix:latest"
)

// NixToolchain implements ToolchainContributor for Nix package provisioning
type NixToolchain struct {
	nix *agentv1alpha1.NixSpec
}

// NewNixToolchain creates a NixToolchain from a NixSpec
func NewNixToolchain(nix *agentv1alpha1.NixSpec) *NixToolchain {
	return &NixToolchain{nix: nix}
}

func (n *NixToolchain) InitContainer() *corev1.Container {
	nixImage := defaultNixImage
	if n.nix.Image != "" {
		nixImage = n.nix.Image
	}

	return &corev1.Container{
		Name:    "nix-env",
		Image:   nixImage,
		Command: []string{"sh"},
		Args:    []string{"-c", n.installScript()},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      nixVolumeName,
				MountPath: "/nix-env",
			},
		},
	}
}

func (n *NixToolchain) Volumes() []corev1.Volume {
	sizeLimit := resource.MustParse("10Gi")
	if n.nix.StoreSizeLimit != "" {
		sizeLimit = resource.MustParse(n.nix.StoreSizeLimit)
	}

	return []corev1.Volume{
		{
			Name: nixVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					SizeLimit: &sizeLimit,
				},
			},
		},
	}
}

func (n *NixToolchain) VolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      nixVolumeName,
			MountPath: nixMountPath,
		},
	}
}

func (n *NixToolchain) EnvVars() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "PATH",
			Value: nixProfileBinPath + ":/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
	}
}

func (n *NixToolchain) installScript() string {
	var pkgRefs []string
	for _, pkg := range n.nix.Packages {
		pkgRefs = append(pkgRefs, "nixpkgs#"+pkg)
	}

	script := "set -e\n"
	script += "cp -a /nix/. /nix-env/\n"
	script += fmt.Sprintf("nix --extra-experimental-features \"nix-command flakes\" profile install --profile /nix/var/nix/profiles/agent %s\n", strings.Join(pkgRefs, " "))
	script += "cp -a /nix/. /nix-env/\n"
	return script
}
