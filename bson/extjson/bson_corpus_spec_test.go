package extjson_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"

	"github.com/skriptble/wilson/bson/extjson"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	Description  string                `json:"description"`
	BsonType     string                `json:"bson_type"`
	TestKey      *string               `json:"test_key"`
	Valid        []validityTestCase    `json:"valid"`
	DecodeErrors []decodeErrorTestCase `json:"decodeErrors"`
	ParseErrors  []parseErrorTestCase  `json:"parseErrors"`
	Deprecated   *bool                 `json:"deprecated"`
}

type validityTestCase struct {
	Description       string  `json:"description"`
	CanonicalBson     string  `json:"canonical_bson"`
	CanonicalExtJson  string  `json:"canonical_extjson"`
	RelaxedExtJson    *string `json:"relaxed_extjson"`
	DegenerateBson    *string `json:"degenerate_bson"`
	DegenerateExtJson *string `json:"degenerate_extjson"`
	ConvertedBson     *string `json:"converted_bson"`
	ConvertedExtJson  *string `json:"converted_extjson"`
	Lossy             *bool   `json:"lossy"`
}

type decodeErrorTestCase struct {
	Description string `json:"description"`
	Bson        string `json:"bson"`
}

type parseErrorTestCase struct {
	Description string `json:"description"`
	String      string `json:"string"`
}

const dataDir = "../../data"

func findJSONFilesInDir(t *testing.T, dir string) []string {
	files := make([]string, 0)

	entries, err := ioutil.ReadDir(dir)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".json" {
			continue
		}

		files = append(files, entry.Name())
	}

	return files
}

func runTest(t *testing.T, file string) {
	filepath := path.Join(dataDir, file)
	content, err := ioutil.ReadFile(filepath)
	require.NoError(t, err)

	// Remove ".json" from filename.
	file = file[:len(file)-5]
	testName := "bson_corpus--builder:" + file

	t.Run(testName, func(t *testing.T) {
		var test testCase
		require.NoError(t, json.Unmarshal(content, &test))

		for _, v := range test.Valid {
			if v.Lossy != nil && *v.Lossy {
				continue
			}

			doc, err := extjson.ParseObjectToBuilder(v.CanonicalExtJson)
			require.NoError(t, err)

			expectedBytes, err := hex.DecodeString(v.CanonicalBson)
			require.NoError(t, err)

			actualBytes := make([]byte, doc.RequiredBytes())
			i, err := doc.WriteDocument(actualBytes)
			require.NoError(t, err)
			require.Len(t, expectedBytes, int(i))
			require.True(t, bytes.Equal(expectedBytes, actualBytes))
		}
	})
}

func Test_BsonCorpus(t *testing.T) {
	for _, file := range findJSONFilesInDir(t, dataDir) {
		runTest(t, file)
	}
}
