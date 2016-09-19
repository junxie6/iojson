# iojson

[![Go Report Card](https://goreportcard.com/badge/github.com/junhsieh/iojson)](https://goreportcard.com/report/github.com/junhsieh/iojson)
[![GoDoc](https://godoc.org/github.com/junhsieh/iojson?status.svg)](https://godoc.org/github.com/junhsieh/iojson)

iojson provides a convenient way to exchange data between your client and server through a uniform JSON format. It helps you to encode data from some Go structs to a JSON string and to decode data from a JSON string to some Go structs. iojson supports storing Go objects to a slice or to a map, which means you could reference your object either by a slice index or by a map key. After populating data from JSON to Go objects, the methods of the objects remained working.

iojson also provides a HTTP middleware function, which works with [Alice](https://github.com/justinas/alice) (a famous middleware chainer).

### How the uniform format looks like?

```
{  
    "Status":true,
    "ErrArr":[],
    "ErrCount":0,
    "ObjArr":[],
    "ObjCount":1,
    "Data":{}
}
```

### Usage

#### Add a object to the slice and the map. Then, let's encode it:

```
type Car struct {
	Name string
}

car := &Car{
	Name: "Init car name",
}

i := iojson.NewIOJSON()
i.AddObj(car)         // add to the slice.
i.AddData("car", car) // add to the map.

fmt.Printf("%s\n", i.Encode()) // encode the data.
```

**Sample output:**

```
{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[{"Name":"Init car name"}],"ObjCount":1,"Data":{"car":{"Name":"Init car name"}}}
```

#### show how to populate data to some existing live objects:

```
package main

import (
	"fmt"
	"strings"
)

import (
	"github.com/junhsieh/iojson"
)

type Car struct {
	Name string
}

func (c *Car) GetName() string {
	return c.Name
}

type House struct {
	Name string
}

func (h *House) GetName() string {
	return h.Name
}

func main() {
	json := `{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[{"Name":"My luxury car"}],"ObjCount":1,"Data":{"house":{"Name":"My luxury house"}}}`

	car := &Car{
		Name: "Init car name",
	}
	house := &House{
		Name: "Init house name",
	}

	i := iojson.NewIOJSON()
	i.AddObj(car) // add car to the slice. Data will be populated after decoded.

	if err := i.Decode(strings.NewReader(json)); err != nil {
		fmt.Printf("err: %s\n", err.Error())
	} else {
		// a live car object.
		fmt.Printf("car: %s\n", car.GetName())

		if v, err := i.GetData("house", house); err != nil {
			fmt.Printf("err: %s\n", err.Error())
		} else {
			// a live house object.
			fmt.Printf("house: %s\n", house.GetName())
			fmt.Printf("house: %s\n", v.(*House).GetName())
		}
	}
}
```

**Sample output:**

```
car: My luxury car
house: My luxury house
house: My luxury house
```

#### This complete example shows converting from JSON to a live Go object through iojson.ObjArr:

```
package main

import (
	"fmt"
	"net/http"
)

import (
	"github.com/junhsieh/iojson"
)

// Car ...
type Car struct {
	Name   string
	Wheels []Wheel
}

// GetName ...
func (c *Car) GetName() string {
	return c.Name
}

// Wheel ...
type Wheel struct {
	Size string
}

// GetSize ...
func (w *Wheel) GetSize() string {
	return w.Size
}

func srvRoot(w http.ResponseWriter, r *http.Request) {
	car := &Car{
		Name: "Init car name",
		Wheels: []Wheel{
			Wheel{Size: "Init wheel size 0"},
			Wheel{Size: "Init wheel size 1"},
		},
	}

	i := iojson.NewIOJSON()
	i.AddObj(car) // car data will be populated once it's decoded.

	if err := i.Decode(r.Body); err != nil {
		w.Write([]byte(err.Error()))
	} else {
		// showing a live object with working methods.
		w.Write([]byte("=========\n"))
		fmt.Fprintf(w, "Car name: %s\n", car.GetName())
		fmt.Fprintf(w, "Wheel size: %s\n", car.Wheels[0].GetSize())
		fmt.Fprintf(w, "Wheel size: %s\n", car.Wheels[1].GetSize())

		// iojson can also encode itself and echo.
		w.Write([]byte("=========\n"))
		i.Echo(w)
		w.Write([]byte("\n"))
	}
}

func main() {
	http.HandleFunc("/", srvRoot)
	http.ListenAndServe(":8080", nil)
}
```

**Run curl command:**

\# curl -H "Content-Type: application/json; charset=UTF-8" -X POST -d '{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[{"Name": "My luxury car","Wheels":[{"Size":"18 inches"},{"Size":"28 inches"}]}],"Data":{}}' http://127.0.0.1:8080/

**Sample outout:**

```
=========
Car name: BMW
Wheel size: 18 inches
Wheel size: 28 inches
=========
{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[{"Name":"BMW","Wheels":[{"Size":"18 inches"},{"Size":"28 inches"}]}],"ObjCount":1,"Data":{}}
```

#### middleware sample:

iojson.EchoHandler stores iojson instance itself in context. Then, run iojson.Echo() at the end of the iojson.EchoHandler through defer function.

```
package main

import (
	"log"
	"net/http"
)

import (
	"github.com/junhsieh/iojson"
	"github.com/justinas/alice"
)

var gMux *http.ServeMux

func srvRoot(w http.ResponseWriter, r *http.Request) {
	o := r.Context().Value(iojson.CTXKey).(*iojson.IOJSON)
	o.AddData("Hello", "World")

	// showing how to add an error message.
	if 1 == 2 {
		o.AddError("my error message")
	}
}

func main() {
	gMux = http.NewServeMux()

	chain := alice.New(
		iojson.EchoHandler, // iojson.EchoHandler stores iojson instance itself in context. Then, run iojson.Echo() at the end of the iojson.EchoHandler through defer function.
	)

	gMux.Handle("/", chain.ThenFunc((srvRoot)))

	srv := &http.Server{
		Handler: gMux,
		Addr:    ":8080",
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("http.ListenAndServe: ", err)
	}
}
```

**Run curl command:**

\# curl -H "Content-Type: application/json; charset=UTF-8" -X GET -d '{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[],"Data":{}}' http://127.0.0.1:8080/

**Sample outout:**

```
{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[],"ObjCount":0,"Data":{"Hello":"World"}}
```
