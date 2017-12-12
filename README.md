# wilson
wilson is a BSON library for Go. It aims to provide types for reading, writing, and manipulating
BSON with low performance and allocation costs. The three main types of this library aim to provide
enough functionality to build more complex logic on top of while also providing benefits of their
own.

## Main Types
The three main types are an append-only builder, a read-only document, and a read-write ordered map.
There is also an Encoder and Decoder that can be used to marshal or unmarshal BSON into one of the
aforementioned types or to a custom type. The Marshaler and Unmarshaler interfaces are provided to
allow types that know how to marshal and unmarshal themselves to avoid reflection during encoding.

### Append-Only Builder
The append-only builder type is called builder.DocumentBuilder. This type allows the
construction of a BSON document incrementally. This type can be used inside of a Marhsaler
implementation by types that wish to avoid reflection.

### Read-Only Document
The read-only document type is called bson.Reader. This type sits directly on top of a byte slice,
allowing fast and low allocation access to the elements of a BSON document and validation of the
document.

### Read-Write Ordered Map
The read-write ordered map type is called bson.Document. This type is the most general and allows
for the construction and modifications of BSON documents.

## Encoder/Decoder
The bson.Encoder type is used to marshal types to an io.Writer. Conversely, the bson.Decoder type is
used to unmarshal an io.Reader into types.

## bson.Marshaler & bson.Unmarshaler
The bson.Marshaler and bson.Unmarshaler types are provided to avoid using reflection when marshaling
or unmarshaling. The Encoder and Decoder will use the methods of these types when available.

## wilson?
bson -> basin -> "binary json" -> \*son -> wildcardson -> wilson
