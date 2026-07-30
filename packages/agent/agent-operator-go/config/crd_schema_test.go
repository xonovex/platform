package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// crdDefinition is the slice of a CustomResourceDefinition this test compares
// against the Go API types.
type crdDefinition struct {
	Spec struct {
		Group string `json:"group"`
		Names struct {
			Kind string `json:"kind"`
		} `json:"names"`
		Versions []struct {
			Name   string `json:"name"`
			Schema struct {
				OpenAPIV3Schema map[string]any `json:"openAPIV3Schema"`
			} `json:"schema"`
		} `json:"versions"`
	} `json:"spec"`
}

// TestCRDSchemasMatchAgentAPITypes reports every field the Go API types declare
// that the structural schema omits, and every schema property no Go field
// backs.
//
// The CRDs under config/crd/bases are maintained by hand, because
// controller-gen is broken on Go 1.25+. A field missing from the schema is
// pruned by the API server before the controller ever sees it, and a property
// no field backs is dropped when the object decodes: both fail silently, and
// TestSamplesStrictlyDecodeAgainstAgentAPI cannot see either, because it
// validates samples against the Go types rather than the schema.
func TestCRDSchemasMatchAgentAPITypes(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := agentv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add agent API to scheme: %v", err)
	}
	paths, err := filepath.Glob(filepath.Join("crd", "bases", "*.yaml"))
	if err != nil {
		t.Fatalf("list CRDs: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no CRD manifests found")
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			var definition crdDefinition
			if err := yaml.Unmarshal(data, &definition); err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			for _, version := range definition.Spec.Versions {
				t.Run(version.Name, func(t *testing.T) {
					gvk := schema.GroupVersionKind{
						Group:   definition.Spec.Group,
						Version: version.Name,
						Kind:    definition.Spec.Names.Kind,
					}
					object, err := scheme.New(gvk)
					if err != nil {
						t.Fatalf("resolve %s: %v", gvk, err)
					}
					compareSchema(
						t,
						definition.Spec.Names.Kind,
						reflect.TypeOf(object).Elem(),
						version.Schema.OpenAPIV3Schema,
					)
				})
			}
		})
	}
}

// jsonMarshaler matches the types that carry their own JSON encoding.
var jsonMarshaler = reflect.TypeOf((*json.Marshaler)(nil)).Elem()

// unexposedUpstreamFields are the fields of embedded Kubernetes types the
// schema deliberately leaves out, keyed by type and JSON name. The API server
// prunes each one, which is the intended behaviour: the operator builds the pod
// itself and cannot honour them.
var unexposedUpstreamFields = map[string]string{
	// Claims name entries of pod.spec.resourceClaims, which an AgentRun cannot declare.
	"v1.ResourceRequirements.claims": "the operator exposes no pod resource claims",
	// FileKeyRef names a volume mount that an AgentRun cannot declare.
	"v1.EnvVarSource.fileKeyRef": "the operator exposes no env-file volume mounts",
}

// compareSchema walks a Go type and its schema node in step, reporting the
// fields and properties that do not line up.
func compareSchema(t *testing.T, path string, goType reflect.Type, node map[string]any) {
	t.Helper()
	// A node that preserves unknown fields keeps everything below it, which is
	// how the CRDs carry embedded Kubernetes types they do not restate.
	if preserved, ok := node["x-kubernetes-preserve-unknown-fields"].(bool); ok && preserved {
		return
	}
	switch goType.Kind() {
	case reflect.Pointer:
		compareSchema(t, path, goType.Elem(), node)
		return
	case reflect.Slice, reflect.Array:
		items, ok := node["items"].(map[string]any)
		if !ok {
			t.Errorf("%s: schema declares no items for a %s", path, goType)
			return
		}
		compareSchema(t, path+"[]", goType.Elem(), items)
		return
	case reflect.Map:
		additional, ok := node["additionalProperties"].(map[string]any)
		if !ok {
			t.Errorf("%s: schema declares no additionalProperties for a %s", path, goType)
			return
		}
		compareSchema(t, path+"{}", goType.Elem(), additional)
		return
	case reflect.Struct:
		// A type with its own JSON encoding, such as a quantity or a
		// timestamp, serialises as a scalar and exposes no properties.
		if reflect.PointerTo(goType).Implements(jsonMarshaler) {
			return
		}
		compareStructSchema(t, path, goType, node)
		return
	default:
		return
	}
}

// compareStructSchema matches a struct's JSON fields against a schema node's
// properties.
func compareStructSchema(t *testing.T, path string, goType reflect.Type, node map[string]any) {
	t.Helper()
	fields := jsonFields(goType)
	if len(fields) == 0 {
		return
	}
	properties, ok := node["properties"].(map[string]any)
	if !ok {
		// ObjectMeta is the one object the API server preserves without a
		// schema; anything else silently loses every field it holds.
		if strings.HasSuffix(path, ".metadata") {
			return
		}
		t.Errorf("%s: schema declares no properties, so every %s field is pruned", path, goType)
		return
	}

	for name, field := range fields {
		property, declared := properties[name].(map[string]any)
		if !declared {
			if _, unexposed := unexposedUpstreamFields[goType.String()+"."+name]; unexposed {
				continue
			}
			t.Errorf("%s.%s: declared by %s but absent from the schema, so the API server prunes it", path, name, goType)
			continue
		}
		compareSchema(t, path+"."+name, field.Type, property)
	}
	for _, name := range sortedKeys(properties) {
		if _, declared := fields[name]; !declared {
			t.Errorf("%s.%s: declared by the schema but no %s field decodes it", path, name, goType)
		}
	}
}

// jsonFields collects the JSON name of every field a struct serialises,
// flattening embedded inline fields into their parent.
func jsonFields(goType reflect.Type) map[string]reflect.StructField {
	fields := map[string]reflect.StructField{}
	for index := range goType.NumField() {
		field := goType.Field(index)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}
		if field.Anonymous && field.Type.Kind() != reflect.Struct {
			continue
		}
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			if field.Anonymous {
				for inlineName, inlineField := range jsonFields(field.Type) {
					fields[inlineName] = inlineField
				}
			}
			continue
		}
		fields[name] = field
	}
	return fields
}

// sortedKeys orders map keys so the reported drift is stable across runs.
func sortedKeys(properties map[string]any) []string {
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
