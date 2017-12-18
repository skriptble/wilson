package bson

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/skriptble/wilson/bson/builder"
	"github.com/skriptble/wilson/bson/extjson"
	"github.com/stretchr/testify/require"
)

var benchmarkDataFiles []string = []string{
	"single_and_multi_document/large_doc.json.gz",
	"single_and_multi_document/small_doc.json.gz",
	"single_and_multi_document/tweet.json.gz",

	"extended_bson/deep_bson.json.gz",
	"extended_bson/flat_bson.json.gz",
	//"extended_bson/full_bson.json.gz",
}

func loadJsonBytesFromFile(filename string) ([]byte, error) {
	compressedData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}

	jsonBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}

func loadDocBuilderFromJsonFile(filename string) (*builder.DocumentBuilder, error) {
	jsonBytes, err := loadJsonBytesFromFile(filename)
	if err != nil {
		return nil, err
	}

	docBuilder, err := extjson.ParseObjectToBuilder(string(jsonBytes))
	if err != nil {
		return nil, err
	}

	return docBuilder, nil
}

func loadFromJsonFile(filename string) (bson.M, bson.D, bson.RawD, *builder.DocumentBuilder, error) {
	docBuilder, err := loadDocBuilderFromJsonFile(filename)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	bsonBytes := make([]byte, docBuilder.RequiredBytes())
	_, err = docBuilder.WriteDocument(bsonBytes)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	bsonM := make(bson.M)
	bsonD := make(bson.D, 0, 8)
	bsonRawD := make(bson.RawD, 0, 8)
	err = bson.Unmarshal(bsonBytes, &bsonM)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = bson.Unmarshal(bsonBytes, &bsonD)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = bson.Unmarshal(bsonBytes, &bsonRawD)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return bsonM, bsonD, bsonRawD, docBuilder, nil
}

type outType int

const (
	bsonM outType = iota
	bsonD
	bsonRawD
	documentBuilder
	extJson
	reader
)

func (ot outType) String() string {
	switch ot {
	case bsonM:
		return "bson.M"
	case bsonD:
		return "bson.D"
	case bsonRawD:
		return "bson.RawD"
	case documentBuilder:
		return "docBuilder"
	case extJson:
		return "extJson"
	case reader:
		return "reader"
	default:
		panic(fmt.Sprintf("Unknown outType. Val: %d", ot))
	}
}

func benchmarkEncodingGen(filename string, out outType) func(b *testing.B) {
	docBsonM, docBsonD, docBsonRawD, docBuilder, err := loadFromJsonFile(filename)

	return func(benchmark *testing.B) {
		if err != nil {
			benchmark.Fatalf("Error parsing file. Filename: %v Err: %v", filename, err)
		}

		switch out {
		case bsonM:
			for idx := 0; idx < benchmark.N; idx++ {
				bson.Marshal(docBsonM)
			}
		case bsonD:
			for idx := 0; idx < benchmark.N; idx++ {
				bson.Marshal(docBsonD)
			}
		case bsonRawD:
			for idx := 0; idx < benchmark.N; idx++ {
				bson.Marshal(docBsonRawD)
			}
		case documentBuilder:
			for idx := 0; idx < benchmark.N; idx++ {
				bsonBytes := make([]byte, docBuilder.RequiredBytes())
				docBuilder.WriteDocument(bsonBytes)
			}
		}
	}
}

func benchmarkDecodingGen(filename string, out outType) (func(b *testing.B), error) {
	docBuilder, err := loadDocBuilderFromJsonFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Error parsing file. Filename: %v Err: %v", filename, err)
	}

	bsonBytes := make([]byte, docBuilder.RequiredBytes())
	_, err = docBuilder.WriteDocument(bsonBytes)
	if err != nil {
		return nil, errors.New("Error writing document to bytes")
	}

	return func(benchmark *testing.B) {
			for idx := 0; idx < benchmark.N; idx++ {
				switch out {
				case bsonM:
					doc := make(bson.M)
					bson.Unmarshal(bsonBytes, &doc)
				case bsonD:
					doc := make(bson.D, 0, 8)
					bson.Unmarshal(bsonBytes, &doc)
				case bsonRawD:
					doc := make(bson.RawD, 0, 8)
					bson.Unmarshal(bsonBytes, &doc)
				case extJson:
					extjson.BsonToExtJson(true, bsonBytes)
				}
			}
		},
		nil
}

func benchmarkRoundtripGen(filename string, out outType) func(b *testing.B) {
	jsonBytes, err := loadJsonBytesFromFile(filename)

	return func(benchmark *testing.B) {
		if err != nil {
			benchmark.Fatalf("Error parsing file. Filename: %v Err: %v", filename, err)
		}

		switch out {
		case documentBuilder:
			for idx := 0; idx < benchmark.N; idx++ {
				doc, err := extjson.ParseObjectToBuilder(string(jsonBytes))
				if err != nil {
					benchmark.Fatal(err)
				}

				bsonBytes := make([]byte, doc.RequiredBytes())
				_, err = doc.WriteDocument(bsonBytes)
				if err != nil {
					benchmark.Fatal(err)
				}

				_, err = extjson.BsonToExtJson(true, bsonBytes)
				if err != nil {
					benchmark.Fatal(err)
				}
			}
		}
	}
}

func benchmarkFirstKeyGen(filename string, out outType) func(benchmark *testing.B) {
	docBuilder, err := loadDocBuilderFromJsonFile(filename)
	var bsonBytes []byte
	if err == nil {
		bsonBytes = make([]byte, docBuilder.RequiredBytes())
		_, err = docBuilder.WriteDocument(bsonBytes)
	}

	return func(b *testing.B) {
		if err != nil {
			b.Fatal(err)
		}

		for i := 0; i < b.N; i++ {
			switch out {
			case bsonM:
				doc := make(bson.M)
				err = bson.Unmarshal(bsonBytes, &doc)
				if err != nil {
					b.Fatal(err)
				}

				var key string
				for key = range doc {
				}

				if len(key) > math.MaxInt32 {
					b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
				}
			case bsonD:
				doc := make(bson.D, 0, 8)
				err = bson.Unmarshal(bsonBytes, &doc)
				if err != nil {
					b.Fatal(err)
				}

				if len(doc[0].Name) > math.MaxInt32 {
					b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
				}
			case bsonRawD:
				doc := make(bson.RawD, 0, 8)
				err = bson.Unmarshal(bsonBytes, &doc)
				if err != nil {
					b.Fatal(err)
				}

				if len(doc[0].Name) > math.MaxInt32 {
					b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
				}
			case reader:
				reader := Reader(bsonBytes)
				elem, err := reader.ElementAt(0)
				require.NoError(b, err)

				if len(elem.Key()) > math.MaxInt32 {
					b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
				}
			}
		}
	}
}

func benchmarkTopLevelKeysGen(filename string, out outType) func(benchmark *testing.B) {
	docBuilder, err := loadDocBuilderFromJsonFile(filename)
	var bsonBytes []byte
	if err == nil {
		bsonBytes = make([]byte, docBuilder.RequiredBytes())
		_, err = docBuilder.WriteDocument(bsonBytes)
	}

	return func(b *testing.B) {
		if err != nil {
			b.Fatal(err)
		}

		for i := 0; i < b.N; i++ {
			switch out {
			case bsonM:
				doc := make(bson.M)
				err = bson.Unmarshal(bsonBytes, &doc)
				if err != nil {
					b.Fatal(err)
				}

				for key := range doc {
					if len(key) > math.MaxInt32 {
						b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
					}
				}
			case bsonD:
				doc := make(bson.D, 0, 8)
				err = bson.Unmarshal(bsonBytes, &doc)
				if err != nil {
					b.Fatal(err)
				}

				for _, elem := range doc {
					if len(elem.Name) > math.MaxInt32 {
						b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
					}
				}
			case bsonRawD:
				doc := make(bson.RawD, 0, 8)
				err = bson.Unmarshal(bsonBytes, &doc)
				if err != nil {
					b.Fatal(err)
				}

				for _, elem := range doc {
					if len(elem.Name) > math.MaxInt32 {
						b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
					}
				}
			case reader:
				reader := Reader(bsonBytes)
				keys, err := reader.Keys(false)
				require.NoError(b, err)

				for _, key := range keys {
					if len(key.Name) > math.MaxInt32 {
						b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
					}
				}
			}
		}
	}
}

func benchmarkAllNestedKeysGen(filename string, out outType) func(benchmark *testing.B) {
	docBuilder, err := loadDocBuilderFromJsonFile(filename)
	var bsonBytes []byte
	if err == nil {
		bsonBytes = make([]byte, docBuilder.RequiredBytes())
		_, err = docBuilder.WriteDocument(bsonBytes)
	}

	return func(b *testing.B) {
		if err != nil {
			b.Fatal(err)
		}

		for i := 0; i < b.N; i++ {
			switch out {
			case bsonM:
				doc := make(bson.M)
				err = bson.Unmarshal(bsonBytes, &doc)
				if err != nil {
					b.Fatal(err)
				}

				assertBsonMKeyLengths(b, doc)
			case bsonD:
				doc := make(bson.D, 0, 8)
				err = bson.Unmarshal(bsonBytes, &doc)
				if err != nil {
					b.Fatal(err)
				}

				assertBsonDKeyLengths(b, doc)
			case bsonRawD:
				doc := make(bson.RawD, 0, 8)
				err = bson.Unmarshal(bsonBytes, &doc)
				if err != nil {
					b.Fatal(err)
				}

				assertBsonRawDKeyLengths(b, doc)
			case reader:
				reader := Reader(bsonBytes)
				keys, err := reader.Keys(true)
				require.NoError(b, err)

				for _, key := range keys {
					if len(key.Name) > math.MaxInt32 {
						b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
					}
				}
			}
		}
	}
}

func assertBsonMKeyLengths(b *testing.B, doc bson.M) {
	for key, val := range doc {
		if len(key) > math.MaxInt32 {
			b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
		}

		assertBsonKeyLengths(b, val)
	}
}

func assertBsonDKeyLengths(b *testing.B, doc bson.D) {
	for _, elem := range doc {
		if len(elem.Name) > math.MaxInt32 {
			b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
		}

		assertBsonKeyLengths(b, elem.Value)
	}
}

func assertBsonRawDKeyLengths(b *testing.B, doc bson.RawD) {
	for _, elem := range doc {
		if len(elem.Name) > math.MaxInt32 {
			b.Fatal("failed unnecessary check to ensure that lookup not optimized out")
		}

		if elem.Value.Kind == 0x03 {
			nestedDoc := make(bson.RawD, 0, 8)
			err := bson.Unmarshal(elem.Value.Data, &nestedDoc)
			if err != nil {
				b.Fatal(err)
			}

			assertBsonRawDKeyLengths(b, nestedDoc)
		}
	}
}

func assertBsonKeyLengths(b *testing.B, doc interface{}) {
	switch d := doc.(type) {
	case bson.M:
		assertBsonMKeyLengths(b, d)
	case bson.D:
		assertBsonDKeyLengths(b, d)
	case bson.RawD:
		assertBsonRawDKeyLengths(b, d)
	}
}

func benchmarkEncoding(benchmark *testing.B) {
	perfBaseDir := "../data/"

	for _, relFilename := range benchmarkDataFiles {
		filename := perfBaseDir + relFilename
		for _, ot := range []outType{bsonM, bsonD, bsonRawD, documentBuilder} {
			benchmark.Run(
				fmt.Sprintf("%v-%v", ot, relFilename),
				benchmarkEncodingGen(filename, ot),
			)
		}
	}
}

func benchmarkDecoding(benchmark *testing.B) {
	perfBaseDir := "../data/"

	for _, relFilename := range benchmarkDataFiles {
		filename := perfBaseDir + relFilename
		for _, ot := range []outType{bsonM, bsonD, bsonRawD, extJson} {
			b, err := benchmarkDecodingGen(filename, ot)
			if err != nil {
				benchmark.Fatal(err)
			}

			benchmark.Run(
				fmt.Sprintf("%v-%v", ot, relFilename),
				b,
			)
		}
	}
}

func benchmarkRoundTrip(benchmark *testing.B) {
	perfBaseDir := "../data/"

	for _, relFilename := range benchmarkDataFiles {
		filename := perfBaseDir + relFilename
		for _, ot := range []outType{documentBuilder} {
			benchmark.Run(
				fmt.Sprintf("%v-%v", ot, relFilename),
				benchmarkRoundtripGen(filename, ot),
			)
		}
	}
}

func BenchmarkFirstKey(benchmark *testing.B) {
	perfBaseDir := "../data/"

	for _, relFilename := range benchmarkDataFiles {
		filename := perfBaseDir + relFilename
		for _, ot := range []outType{bsonM, bsonD, bsonRawD, reader} {
			benchmark.Run(
				fmt.Sprintf("%v-%v", ot, relFilename),
				benchmarkFirstKeyGen(filename, ot),
			)
		}
	}
}

func BenchmarkTopLevelKeys(benchmark *testing.B) {
	perfBaseDir := "../data/"

	for _, relFilename := range benchmarkDataFiles {
		filename := perfBaseDir + relFilename
		for _, ot := range []outType{bsonM, bsonD, bsonRawD, reader} {
			benchmark.Run(
				fmt.Sprintf("%v-%v", ot, relFilename),
				benchmarkTopLevelKeysGen(filename, ot),
			)
		}
	}
}

func BenchmarkAllNestedKeys(benchmark *testing.B) {
	perfBaseDir := "../data/"

	for _, relFilename := range benchmarkDataFiles {
		filename := perfBaseDir + relFilename
		for _, ot := range []outType{bsonM, bsonD, bsonRawD, reader} {
			benchmark.Run(
				fmt.Sprintf("%v-%v", ot, relFilename),
				benchmarkAllNestedKeysGen(filename, ot),
			)
		}
	}
}

// Asserts each test file can be read from disk and parsed into the bson data structures.
func TestTest(test *testing.T) {
	perfBaseDir := "../data/"

	for _, relFilename := range benchmarkDataFiles {
		filename := perfBaseDir + relFilename
		bsonM, bsonD, bsonRawD, docBuilder, err := loadFromJsonFile(filename)
		if err != nil {
			test.Fatalf("Error parsing file. Filename: %v Err: %v", filename, err)
		}

		_, _, _, _ = bsonM, bsonD, bsonRawD, docBuilder

		// fmt.Println(bsonM)
		// fmt.Println(bsonD)
		// fmt.Println(bsonRawD)
	}
}
