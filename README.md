# iojson

[![Go Report Card](https://goreportcard.com/badge/github.com/junhsieh/iojson)](https://goreportcard.com/report/github.com/junhsieh/iojson)
[![GoDoc](https://godoc.org/github.com/junhsieh/iojson?status.svg)](https://godoc.org/github.com/junhsieh/iojson)

iojson provides a convenient way to exchange data between your client and server through a uniform JSON format. It helps you to encode data from some Go structs to a JSON string and to decode data from a JSON string to some Go structs. iojson supports storing Go objects to a slice or to a map, which means you could reference your object either by a slice index or by a map key. After populating data from JSON to Go objects, the methods of the objects remained working.

iojson also provides a HTTP middleware function, which works with a famous middleware chainer called [Alice](https://github.com/justinas/alice).

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

#### Add a object to the slice and the map then encode:

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

fmt.Printf("%s\n", i.Encode())
```

**Sample output:**

{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[{"Name":"Init car name"}],"ObjCount":1,"Data":{"car":{"Name":"Init car name"}}}

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

\# curl -H "Content-Type: application/json; charset=UTF-8" -X GET -d '{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[{"Name": "BMW","Wheels":[{"Size":"18 inches"},{"Size":"28 inches"}]}],"Data":{}}' http://127.0.0.1:8080/
