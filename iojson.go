package iojson

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

// Size constants
const (
	KB = 1 << 10
	MB = 1 << 20
)

const (
	// CKey is the key name for context
	CKey = "iojsonCkey"
)

var (
	// IOLimitReaderSize ...
	IOLimitReaderSize int64 = 2 * MB
)

// D ...
type D map[string]interface{}

// IOJSON ...
type IOJSON struct {
	Status       bool
	ErrArr       []string
	ErrCount     int8
	ObjArr       []interface{}
	sync.RWMutex // embedded.  see http://golang.org/ref/spec#Struct_types
	Data         D
}

// NewIOJSON ...
func NewIOJSON() *IOJSON {
	return &IOJSON{
		ErrArr: []string{},
		ObjArr: []interface{}{},
		Data:   make(D),
	}
}

// AddError ...
func (o *IOJSON) AddError(str string) {
	o.ErrArr = append(o.ErrArr, str)
	o.ErrCount++
}

// AddObj ...
func (o *IOJSON) AddObj(v interface{}) {
	o.ObjArr = append(o.ObjArr, v)
}

// AddData ...
func (o *IOJSON) AddData(k string, v interface{}) {
	o.Lock()
	defer o.Unlock()
	o.Data[k] = v
}

// GetData ...
func (o *IOJSON) GetData(k string) interface{} {
	o.RLock()
	defer o.RUnlock()
	return o.Data[k]
}

// JSONFail ...
func (o *IOJSON) JSONFail(err error) string {
	// TODO: propery way to escape characters in err.Error()?
	return `{"Status":false,"ErrArr":["` + strings.Replace(err.Error(), `"`, ``, -1) + `"],"ErrCount":1}`
}

// Encode encodes the object itself to JSON and return []byte.
func (o *IOJSON) Encode() []byte {
	// TODO: find out the difference between the following three lines and io.Reader.
	//var b bytes.Buffer
	//bytes.NewBuffer([]byte("test"))
	b := new(bytes.Buffer)

	if o.ErrCount == 0 {
		o.Status = true
	} else {
		// reset to default
		o.Status = false
		o.ObjArr = []interface{}{}
		o.Data = make(D)
	}

	if err := json.NewEncoder(b).Encode(o); err != nil {
		return []byte(o.JSONFail(err))
	}

	return b.Bytes()
}

// EncodePretty ...
func (o *IOJSON) EncodePretty() []byte {
	var b bytes.Buffer

	if err := json.Indent(&b, o.Encode(), "", "  "); err != nil {
		return []byte(o.JSONFail(err))
	}

	return b.Bytes()
}

// EncodeString encodes the object itself to JSON and return string.
func (o *IOJSON) EncodeString() string {
	return string(o.Encode())
}

// Decode ...
func (o *IOJSON) Decode(b io.Reader) error {
	if err := json.NewDecoder(io.LimitReader(b, IOLimitReaderSize)).Decode(o); err != nil {
		return err
	}

	return nil
}

// Echo ...
func (o *IOJSON) Echo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if _, err := w.Write(o.Encode()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// EchoHandler ...
func EchoHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("DEBUG: iojson.EchoHandler: Inside")

		o := NewIOJSON()
		ctx := context.WithValue(r.Context(), CKey, o)

		defer func() {
			o.Echo(w)
		}()

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ErrorHandler ...
func ErrorHandler(errstr string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: iojson.ErrorHandler: Inside")

		o := r.Context().Value(CKey).(*IOJSON)

		if errstr != "" {
			o.AddError(errstr)
		} else {
			o.AddError("iojson.ErrorHandler")
		}
	})
}
