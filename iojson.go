// Package iojson ...
package iojson

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
)

var debugLineNumPtr = flag.Bool("debugLineNum", false, "show filename and line number")

// Size constants
const (
	KB = 1 << 10
	MB = 1 << 20
)

const (
	// CTXKey is the key name for context
	CTXKey = "iojsonCTXKey"
)

const (
	// ErrKeyNotExist ...
	ErrKeyNotExist = " key does not exist"
	// ErrJSONRawIsNil ...
	ErrJSONRawIsNil = "jsonRaw is nil"
)

var (
	// IOLimitReaderSize ...
	IOLimitReaderSize int64 = 2 * MB
)

// JSONRawArr uses *json.RawMessage (instead of interface{}) to delay JSON decoding until we supplied an object.
type JSONRawArr []*json.RawMessage

// JSONRawMap ...
type JSONRawMap map[string]*json.RawMessage

// IOJSON ...
type IOJSON struct {
	Status       bool
	ErrArr       []string
	ObjArr       JSONRawArr // NOTE: do not access this field directly.
	sync.RWMutex            // embedded. see http://golang.org/ref/spec#Struct_types
	ObjMap       JSONRawMap // NOTE: do not access this field directly.
}

// NewIOJSON ...
func NewIOJSON() *IOJSON {
	return &IOJSON{
		ErrArr: []string{},
		ObjArr: make(JSONRawArr, 0),
		ObjMap: make(JSONRawMap),
	}
}

// AddError ...
func (o *IOJSON) AddError(str string) {
	o.ErrArr = append(o.ErrArr, str)

	if *debugLineNumPtr == true {
		pc, fn, line, _ := runtime.Caller(1)
		o.ErrArr = append(o.ErrArr, runtime.FuncForPC(pc).Name()+"["+fn+":"+strconv.Itoa(line)+"]")
	}
}

// AddObjToArr ...
func (o *IOJSON) AddObjToArr(v interface{}) error {
	var b []byte
	var err error

	if b, err = json.Marshal(v); err != nil {
		return err
	}

	o.ObjArr = append(o.ObjArr, o.NewRawMessage(b))
	return nil
}

// GetObjFromArr ...
func (o *IOJSON) GetObjFromArr(k int, obj interface{}) (interface{}, error) {
	if k < 0 || k >= len(o.ObjArr) {
		return nil, errors.New(strconv.Itoa(k) + ErrKeyNotExist)
	}

	// NOTE: the primitive types (int, string) will not work if obj (instead of &obj) is used.
	return obj, o.populateObj(o.ObjArr[k], &obj)
}

// AddObjToMap ...
func (o *IOJSON) AddObjToMap(k string, v interface{}) error {
	o.Lock()
	defer o.Unlock()

	var b []byte
	var err error

	if b, err = json.Marshal(v); err != nil {
		return err
	}

	o.ObjMap[k] = o.NewRawMessage(b)

	return nil
}

// GetObjFromMap ...
func (o *IOJSON) GetObjFromMap(k string, obj interface{}) (interface{}, error) {
	o.RLock()
	defer o.RUnlock()

	var jsonRaw *json.RawMessage
	var ok bool

	if jsonRaw, ok = o.ObjMap[k]; !ok {
		return nil, errors.New(k + ErrKeyNotExist)
	}

	// NOTE: the primitive types (int, string) will not work if obj (instead of &obj) is used.
	return obj, o.populateObj(jsonRaw, &obj)
}

// NewRawMessage ...
func (o *IOJSON) NewRawMessage(b []byte) *json.RawMessage {
	j := json.RawMessage(b)
	return &j

	// NOTE: another way of assigning value to *json.RawMessage.
	//jPtr := new(json.RawMessage)
	//*jPtr = b
	//return jPtr
}

// populateObj ...
func (o *IOJSON) populateObj(jsonRaw *json.RawMessage, obj interface{}) error {
	if jsonRaw == nil {
		return errors.New(ErrJSONRawIsNil)
	}

	if err := json.NewDecoder(bytes.NewReader(*jsonRaw)).Decode(obj); err != nil {
		return err
	}

	return nil
}

// JSONFail ...
func (o *IOJSON) JSONFail(err error) string {
	log.Printf("err: %v", err.Error())
	return `{"Status":false,"ErrArr":["Encode failed. Check log"],"ObjArr":[],"ObjMap":{}}`
}

// Encode encodes the object itself to JSON and return []byte.
func (o *IOJSON) Encode() []byte {
	if len(o.ErrArr) > 0 {
		// reset to default
		o.Status = false
		o.ObjArr = make(JSONRawArr, 0)
		o.ObjMap = make(JSONRawMap)
	} else {
		o.Status = true
	}

	//
	var b []byte
	var err error

	if b, err = json.Marshal(o); err != nil {
		return []byte(o.JSONFail(err))
	}

	return b
}

// EncodePretty ...
func (o *IOJSON) EncodePretty() []byte {
	var b bytes.Buffer

	if err := json.Indent(&b, o.Encode(), "", "    "); err != nil {
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

		// TODO: should use use basic type string as key in context.WithValue:
		// https://medium.com/@matryer/context-keys-in-go-5312346a868d#.kjr3hh7nh
		ctx := context.WithValue(r.Context(), CTXKey, o)

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

		o := r.Context().Value(CTXKey).(*IOJSON)

		if errstr != "" {
			o.AddError(errstr)
		} else {
			o.AddError("iojson.ErrorHandler")
		}
	})
}
