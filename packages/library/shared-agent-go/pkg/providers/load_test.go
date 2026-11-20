package providers

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

type mapReader struct {
	files        fstest.MapFS
	readDirErr   error
	readFileErr  error
	readFileFail string
}

func (m mapReader) ReadDir(dir string) ([]fs.DirEntry, error) {
	if m.readDirErr != nil {
		return nil, m.readDirErr
	}
	return fs.ReadDir(m.files, strings.TrimPrefix(dir, "/"))
}

func (m mapReader) ReadFile(file string) ([]byte, error) {
	if m.readFileErr != nil && (m.readFileFail == "" || strings.HasSuffix(file, m.readFileFail)) {
		return nil, m.readFileErr
	}
	return fs.ReadFile(m.files, strings.TrimPrefix(file, "/"))
}

func TestLoadDirReadsEachProviderFileInNameOrder(t *testing.T) {
	reader := mapReader{files: fstest.MapFS{
		"p/b.json":     {Data: []byte(`{"name":"b","agent":"claude"}`)},
		"p/a.json":     {Data: []byte(`{"name":"a","agent":"codex"}`)},
		"p/notes.txt":  {Data: []byte("ignored")},
		"p/sub/c.json": {Data: []byte(`{"name":"c","agent":"claude"}`)},
	}}

	documents, err := LoadDir(reader, "/p", "user")

	if err != nil || len(documents) != 2 {
		t.Fatalf("LoadDir() = %+v, %v", documents, err)
	}
	if documents[0].Preset.Name != "a" || documents[1].Preset.Name != "b" {
		t.Errorf("documents = %+v, want name order", documents)
	}
	if documents[0].Layer != "user" || documents[0].Origin != "/p/a.json" || string(documents[0].Raw) != `{"name":"a","agent":"codex"}` {
		t.Errorf("document = %+v", documents[0])
	}
}

func TestLoadDirOfAMissingDirectoryHoldsNothing(t *testing.T) {
	documents, err := LoadDir(mapReader{files: fstest.MapFS{}}, "/absent", "user")
	if err != nil || documents != nil {
		t.Errorf("LoadDir() = %+v, %v", documents, err)
	}
}

func TestLoadDirFailsClosedAndNamesTheFile(t *testing.T) {
	failure := errors.New("denied")
	cases := map[string]struct {
		reader mapReader
		phrase string
	}{
		"directory": {mapReader{readDirErr: failure}, "read providers directory /p"},
		"file":      {mapReader{files: fstest.MapFS{"p/a.json": {Data: []byte("{}")}}, readFileErr: failure}, "read provider file /p/a.json"},
		"invalid":   {mapReader{files: fstest.MapFS{"p/a.json": {Data: []byte(`{"name":"a"}`)}}}, "provider file /p/a.json"},
		"mismatch":  {mapReader{files: fstest.MapFS{"p/a.json": {Data: []byte(`{"name":"b","agent":"claude"}`)}}}, "name the file b.json"},
	}
	for name, test := range cases {
		if _, err := LoadDir(test.reader, "/p", "user"); err == nil || !strings.Contains(err.Error(), test.phrase) {
			t.Errorf("%s: error = %v, want %q", name, err, test.phrase)
		}
	}
}
