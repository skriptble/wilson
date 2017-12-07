package bson

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"testing"

	"errors"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/skriptble/wilson/bson/extjson"
	"github.com/skriptble/wilson/builder"
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

func BenchmarkRoundTrip(benchmark *testing.B) {
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
