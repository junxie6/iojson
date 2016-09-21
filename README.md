# iojson

[![Go Report Card](https://goreportcard.com/badge/github.com/junhsieh/iojson)](https://goreportcard.com/report/github.com/junhsieh/iojson)
[![GoDoc](https://godoc.org/github.com/junhsieh/iojson?status.svg)](https://godoc.org/github.com/junhsieh/iojson)

iojson provides a convenient way to exchange data between your client and server through a uniform JSON format. It helps you to encode data from Go structs to a JSON string and to decode data from a JSON string to Go structs. iojson supports storing Go objects to a slice or to a map, which means you could reference your object either by a slice index or by a map key according to your preference. After populating data from JSON to Go objects, the methods of the objects remained working.

iojson also provides a HTTP middleware function, which works with [Alice](https://github.com/justinas/alice) (a famous middleware chainer).

## How the uniform format looks like?

```
{
    "Status": true,
    "ErrArr": [],
    "ErrCount": 0,
    "ObjArr": [],
    "ObjMap": {}
}
```

## Usage

### packages and structure definitions for examples:

```
import (
	"fmt"
	"github.com/junhsieh/iojson"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func NewCar() *Car {
	return &Car{ItemArr: make([]Item, 0)}
}

type Car struct {
	Name    string
	ItemArr []Item
}

func (c *Car) GetName() string {
	return c.Name
}

func NewItem() *Item {
	return &Item{}
}

type Item struct {
	Name string
}

func (i *Item) GetName() string {
	return i.Name
}
```

### Add an object to the slice. Then, encode it:

```
func main() {
	item := NewItem()
	item.Name = "Bag"

	car := NewCar()
	car.Name = "My luxury car"
	car.ItemArr = append(car.ItemArr, *item)

	o := iojson.NewIOJSON()
	o.AddObjToArr(car) // add the car object to the slice.

	fmt.Printf("%s\n", o.EncodePretty()) // encode data with nice format or call o.Encode().
}
```

**Sample output:**

```
{
    "Status": true,
    "ErrArr": [],
    "ErrCount": 0,
    "ObjArr": [
        {
            "Name": "My luxury car",
            "ItemArr": [
                {
                    "Name": "Bag"
                }
            ]
        }
    ],
    "ObjMap": {}
}
```

### Add an object to the map. Then, encode it:

```
func main() {
	item := NewItem()
	item.Name = "Bag"

	car := NewCar()
	car.Name = "My luxury car"
	car.ItemArr = append(car.ItemArr, *item)

	o := iojson.NewIOJSON()
	o.AddObjToMap("Car", car) // add the car object to the map.

	fmt.Printf("%s\n", o.EncodePretty()) // encode data with nice format or call o.Encode().
}
```

**Sample output:**

```
{
    "Status": true,
    "ErrArr": [],
    "ErrCount": 0,
    "ObjArr": [],
    "ObjMap": {
        "Car": {
            "Name": "My luxury car",
            "ItemArr": [
                {
                    "Name": "Bag"
                }
            ]
        }
    }
}
```

### Show how to populate data to some existing live objects:

```
func main() {
	jsonStr := `{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[{"Name":"My luxury car","ItemArr":[{"Name":"Bag"},{"Name":"Pen"}]}],"ObjMap":{}}`

	car := NewCar()

	i := iojson.NewIOJSON()

	if err := i.Decode(strings.NewReader(jsonStr)); err != nil {
		fmt.Printf("err: %s\n", err.Error())
	}

	// populating data to a live car object.
	if v, err := i.GetObjFromArr(0, car); err != nil {
		fmt.Printf("err: %s\n", err.Error())
	} else {
		fmt.Printf("car (original): %s\n", car.GetName())
		fmt.Printf("car (returned): %s\n", v.(*Car).GetName())

		for k, item := range car.ItemArr {
			fmt.Printf("ItemArr[%d] of car (original): %s\n", k, item.GetName())
		}

		for k, item := range v.(*Car).ItemArr {
			fmt.Printf("ItemArr[%d] of car (returned): %s\n", k, item.GetName())
		}
	}
}
```

**Sample output:**

```
car (original): My luxury car
car (returned): My luxury car
ItemArr[0] of car (original): Bag
ItemArr[1] of car (original): Pen
ItemArr[0] of car (returned): Bag
ItemArr[1] of car (returned): Pen
```

### This complete example shows how to use iojson with HTTP handler:

```
func srvRoot(w http.ResponseWriter, r *http.Request) {
	car := NewCar()

	i := iojson.NewIOJSON() // for input
	o := iojson.NewIOJSON() // for output

	if err := i.Decode(r.Body); err != nil {
		o.AddError(err.Error())
		o.Echo(w)
		return
	}

	// populating data to a live car object.
	if _, err := i.GetObjFromMap("Car", car); err != nil {
		o.AddError(err.Error())
		o.Echo(w)
		return
	}

	o.AddObjToMap("Car.Name", car.GetName())

	for k, item := range car.ItemArr {
		o.AddObjToMap("Car.ItemArr["+strconv.Itoa(k)+"]", item.GetName())
	}

	o.Echo(w)
}

func main() {
	http.HandleFunc("/", srvRoot)
	http.ListenAndServe(":8080", nil)
}
```

**Run curl command:**

\# curl -H "Content-Type: application/json; charset=UTF-8" -X POST -d '{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[],"ObjMap":{"Car":{"Name":"My luxury car","ItemArr":[{"Name":"Bag"},{"Name":"Pen"}]}}}' http://127.0.0.1:8080/

**Sample outout:**

```
{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[],"ObjMap":{"Car.ItemArr[0]":"Bag","Car.ItemArr[1]":"Pen","Car.Name":"My luxury car"}}
```

### using the middleware provided by iojson example:

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
	o.AddObjToMap("Hello", "World")

	// NOTE: do not call o.Echo(w) if you are using iojson.EchoHandler middleware. Because it will call it for you.
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

\# curl -H "Content-Type: application/json; charset=UTF-8" -X GET http://127.0.0.1:8080/

**Sample outout:**

```
{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[],"ObjMap":{"Hello":"World"}}
```
