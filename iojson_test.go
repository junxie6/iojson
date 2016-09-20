package iojson

import (
	"fmt"
	//"reflect"
	"strings"
	"testing"
)

type Car struct {
	Name    string
	ItemArr []Item
}

func (c *Car) GetName() string {
	return c.Name
}

type Item struct {
	Name string
}

func (i *Item) GetName() string {
	return i.Name
}

type TestCase struct {
	json        string
	key         string
	keyNotExist string
	want        string
	obj         interface{}
}

func GetTestCase(storeType string) []TestCase {
	return []TestCase{
		{`{"` + storeType + `":{"%s":{"Name":"%s"}}}`, "Car", "", "My luxury car", &Car{Name: "Init Car"}},
		{`{"` + storeType + `":{"%s":{"Name":"%s"}}}`, "Car", "Dummy", "My luxury car", &Car{Name: "Init Car"}},
		{`{"` + storeType + `":{"%s":"%s"}}`, "Hello", "", "World", nil},
		{`{"` + storeType + `":{"%s":"%s"}}`, "Hello", "Dummy", "World", nil},
		{`{"` + storeType + `":{"%s":%s}}`, "Amt", "", "123.8", nil},
		{`{"` + storeType + `":{"%s":%s}}`, "Amt", "Dummy", "123.8", nil},
	}
}

func TestGetData(t *testing.T) {
	var tests = GetTestCase("Data")

	for _, test := range tests {
		//fmt.Printf("HERE: %v\n", reflect.TypeOf(test.obj))
		//theType := reflect.New(reflect.TypeOf(test.obj)).Interface()

		test.json = fmt.Sprintf(test.json, test.key, test.want)
		//fmt.Printf("HERE: %v\n", test.json)

		test.key += test.keyNotExist

		i := NewIOJSON()

		if err := i.Decode(strings.NewReader(test.json)); err != nil {
			t.Errorf("i.Decode(strings.NewReader(%v)) = %v", test.json, err)

			continue
		}

		if val, err := i.GetData(test.key, test.obj); err != nil {
			if err.Error() == test.key+ErrDataKeyNotExist {
				// Do nothing. Recognized error.
				fmt.Printf("%v (not exist): %#v\n", test.key, val)
			} else {
				t.Errorf("i.GetData(%v, %v) = %v", test.key, test.obj, err)
			}

			continue
		} else {
			switch v := test.obj.(type) {
			case *Car:
				// use the original object.
				if name := test.obj.(*Car).GetName(); name != test.want {
					t.Errorf("%v.GetName() = %v; want = %v", test.key, name, test.want)
				} else {
					fmt.Printf("%v: %#v (original object)\n", test.key, name)
				}

				// use the returned object.
				if name := val.(*Car).GetName(); name != test.want {
					t.Errorf("%v.GetName() = %v; want = %v", test.key, name, test.want)
				} else {
					fmt.Printf("%v: %#v (returned object)\n", test.key, name)
				}
			case nil:
				fmt.Printf("%v: %#v\n", test.key, val)
			default:
				t.Errorf("test.obj(type) = %v", v)
			}
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	o := NewIOJSON()

	o.AddData("test", "test")

	for i := 0; i < b.N; i++ {
		o.Encode()
	}
}

func BenchmarkAddData(b *testing.B) {
	for i := 0; i < b.N; i++ {
		car := &Car{
			Name: "Init Car",
		}
		o := NewIOJSON()
		o.AddData("Car", car)
		o.AddData("Hello", "World")
		o.AddData("Age", 18)
		o.Encode()
	}
}
