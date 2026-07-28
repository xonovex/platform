package v1alpha1

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

// maxFillDepth bounds the walk so a self-referential type cannot recurse
// forever. No API type nests anywhere near this deep.
const maxFillDepth = 8

// retainedRootTypes is every type AddToScheme registers, which is the set whose
// deep copies the API server relies on.
func retainedRootTypes() []runtime.Object {
	return []runtime.Object{
		&AgentHarness{}, &AgentHarnessList{},
		&AgentPolicy{}, &AgentPolicyList{},
		&AgentProvider{}, &AgentProviderList{},
		&AgentRun{}, &AgentRunList{},
		&AgentToolchain{}, &AgentToolchainList{},
		&AgentWorkspace{}, &AgentWorkspaceList{},
	}
}

// TestDeepCopyLeavesNoSharedReferences populates every pointer, slice and map
// reachable from each root type, deep copies it, and reports any reference the
// copy still shares with its source.
//
// zz_generated.deepcopy.go carries a "DO NOT EDIT" marker but is maintained by
// hand, because controller-gen omits subtype DeepCopyInto methods on Go 1.25+.
// A field added to a spec without a matching line in DeepCopyInto produces
// exactly this aliasing, and no other test in the package would see it:
// TestRetainedTypesSupportRuntimeDeepCopy copies zero-value objects, where every
// reference is nil and nothing can alias.
func TestDeepCopyLeavesNoSharedReferences(t *testing.T) {
	for _, object := range retainedRootTypes() {
		objectType := reflect.TypeOf(object).Elem()
		t.Run(objectType.Name(), func(t *testing.T) {
			source := reflect.New(objectType)
			fillReferences(source.Elem(), 0)

			original, ok := source.Interface().(runtime.Object)
			if !ok {
				t.Fatalf("%s does not implement runtime.Object", objectType.Name())
			}
			copied := reflect.ValueOf(original.DeepCopyObject())

			for _, path := range sharedReferences(
				source.Elem(), copied.Elem(), objectType.Name(),
			) {
				t.Errorf("deep copy shares %s with its source", path)
			}
		})
	}
}

// fillReferences gives every reference-typed field a non-nil value, so that a
// DeepCopyInto that forgets one has something observable to share.
func fillReferences(value reflect.Value, depth int) {
	if depth > maxFillDepth || !value.CanSet() {
		return
	}
	switch value.Kind() {
	case reflect.Pointer:
		value.Set(reflect.New(value.Type().Elem()))
		fillReferences(value.Elem(), depth+1)
	case reflect.Slice:
		value.Set(reflect.MakeSlice(value.Type(), 1, 1))
		fillReferences(value.Index(0), depth+1)
	case reflect.Map:
		value.Set(reflect.MakeMap(value.Type()))
		key := reflect.New(value.Type().Key()).Elem()
		fillReferences(key, depth+1)
		entry := reflect.New(value.Type().Elem()).Elem()
		fillReferences(entry, depth+1)
		value.SetMapIndex(key, entry)
	case reflect.Struct:
		for field := range value.NumField() {
			fillReferences(value.Field(field), depth+1)
		}
	case reflect.String:
		value.SetString("x")
	case reflect.Bool:
		value.SetBool(true)
	case reflect.Int, reflect.Int32, reflect.Int64:
		value.SetInt(1)
	default:
	}
}

// sharedReferences reports the paths at which source and copy still point at the
// same pointer, slice backing array or map.
func sharedReferences(source, copied reflect.Value, path string) []string {
	if source.Kind() != copied.Kind() {
		return nil
	}
	switch source.Kind() {
	case reflect.Pointer:
		if source.IsNil() || copied.IsNil() {
			return nil
		}
		if source.Pointer() == copied.Pointer() {
			return []string{path + " (pointer)"}
		}
		return sharedReferences(source.Elem(), copied.Elem(), path)
	case reflect.Slice:
		if source.IsNil() || copied.IsNil() || source.Len() == 0 {
			return nil
		}
		if source.Pointer() == copied.Pointer() {
			return []string{path + " (slice)"}
		}
		var shared []string
		for index := range source.Len() {
			shared = append(shared, sharedReferences(
				source.Index(index), copied.Index(index), path+"[0]",
			)...)
		}
		return shared
	case reflect.Map:
		if source.IsNil() || copied.IsNil() {
			return nil
		}
		if source.Pointer() == copied.Pointer() {
			return []string{path + " (map)"}
		}
		return nil
	case reflect.Struct:
		var shared []string
		for field := range source.NumField() {
			if !source.Field(field).CanInterface() {
				continue
			}
			shared = append(shared, sharedReferences(
				source.Field(field),
				copied.Field(field),
				path+"."+source.Type().Field(field).Name,
			)...)
		}
		return shared
	default:
		return nil
	}
}
