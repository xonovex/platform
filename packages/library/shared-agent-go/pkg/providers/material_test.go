package providers

import (
	"reflect"
	"testing"
)

const projectRoute = `{"name":"proj","agent":"claude","models":{"big":{"id":"big-id"},"small":{"id":"small-id"}},"environment":{"BASE_URL":"https://models.example.test:8443/v1","OTHER":"http://127.0.0.1:8317/v1","NOTE":"no url"},"args":["-c","base_url=\"https://args.example.test/x\""]}`

func TestMaterialDigestChangesWithTheFileAndTheModels(t *testing.T) {
	base := Material{Documents: []Document{doc(t, "project", "/r/.agent-cli/providers/proj.json", projectRoute)}, Models: []string{"big=x"}}
	digest := base.Digest()
	if digest == "" {
		t.Fatal("Digest() = empty")
	}
	if base.Digest() != digest {
		t.Error("Digest() is not stable")
	}

	edited := base
	edited.Documents = []Document{doc(t, "project", "/r/.agent-cli/providers/proj.json", projectRoute+"\n")}
	if edited.Digest() == digest {
		t.Error("a changed file kept the digest")
	}
	remodeled := base
	remodeled.Models = []string{"big=y"}
	if remodeled.Digest() == digest {
		t.Error("a changed model override kept the digest")
	}
	withoutModels := Material{Documents: base.Documents}
	if withoutModels.Digest() == digest {
		t.Error("dropping the model overrides kept the digest")
	}
}

func TestMaterialEmpty(t *testing.T) {
	if !(Material{}).Empty() {
		t.Error("Material{} is not empty")
	}
	if (Material{Models: []string{"a=b"}}).Empty() || (Material{Documents: []Document{{}}}).Empty() {
		t.Error("material with content is empty")
	}
}

func TestMaterialDescribeShowsHostsAndSlots(t *testing.T) {
	m := Material{Documents: []Document{doc(t, "project", "/r/proj.json", projectRoute)}}

	got := m.Describe()

	want := []Description{{
		Name: "proj", Agent: "claude", Origin: "/r/proj.json",
		Hosts: []string{"127.0.0.1:8317", "args.example.test", "models.example.test:8443"},
		Slots: []string{"big=big-id", "small=small-id"},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Describe() = %+v, want %+v", got, want)
	}
}

func TestMaterialDigestOfUnconfinedPathsIsEmpty(t *testing.T) {
	bad := Document{Preset: Preset{Name: "../x", Agent: "claude"}}
	if digest := (Material{Documents: []Document{bad}}).Digest(); digest != "" {
		t.Errorf("Digest() = %q, want empty for a name that escapes", digest)
	}
}
