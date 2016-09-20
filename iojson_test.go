package iojson

import (
	"fmt"
	"reflect"
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
	want2       string
	obj         interface{}
}

func GetTestCase(storeType string) []TestCase {
	return []TestCase{
		//{`{"` + storeType + `":{"%s":{"Name":"%s","ItemArr":[{"Name":"%s"}]}}}`, "Car", "", "My luxury car", "Bag", &Car{}},
		{`{"` + storeType + `":{"%s":[{"Name":"%s","ItemArr":[{"Name":"%s"}]}]}}`, "Car", "", "My luxury car", "Bag", &[]Car{}},
		//{`{"` + storeType + `":{"%s":"%s"}}`, "Hello", "", "World", "", nil},
		//{`{"` + storeType + `":{"%s":"%s"}}`, "Hello", "Dummy", "World", "", nil},
		//{`{"` + storeType + `":{"%s":%s}}`, "Amt", "", "123.8", "", nil},
		//{`{"` + storeType + `":{"%s":%s}}`, "Amt", "Dummy", "123.8", "", nil},
	}
}

func TestGetData(t *testing.T) {
	var tests = GetTestCase("Data")

	for _, test := range tests {
		//fmt.Printf("HERE: %v\n", reflect.TypeOf(test.obj))
		//theType := reflect.New(reflect.TypeOf(test.obj)).Interface()

		test.json = fmt.Sprintf(test.json, test.key, test.want, test.want2)
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
					t.Errorf("%v.GetName() = %v; want = %v", reflect.TypeOf(test.obj.(*Car)), test.obj.(*Car).GetName(), test.want)
				} else if name := test.obj.(*Car).ItemArr[0].GetName(); name != test.want2 {
					t.Errorf("%v.GetName() = %v; want = %v", reflect.TypeOf(test.obj.(*Car).ItemArr[0]), test.obj.(*Car).ItemArr[0].GetName(), test.want2)
				} else {
					fmt.Printf("%v (original object): %#v\n", reflect.TypeOf(test.obj.(*Car)), test.obj.(*Car).GetName())
					fmt.Printf("%v (original object): %#v\n", reflect.TypeOf(test.obj.(*Car).ItemArr[0]), test.obj.(*Car).ItemArr[0].GetName())
				}

				// use the returned object.
				if name := val.(*Car).GetName(); name != test.want {
					t.Errorf("%v.GetName() = %v; want = %v", reflect.TypeOf(val.(*Car)), val.(*Car).GetName(), test.want)
				} else if name := val.(*Car).ItemArr[0].GetName(); name != test.want2 {
					t.Errorf("%v.GetName() = %v; want = %v", reflect.TypeOf(test.obj.(*Car).ItemArr[0]), val.(*Car).ItemArr[0].GetName(), test.want2)
				} else {
					fmt.Printf("%v (returned object): %#v\n", reflect.TypeOf(val.(*Car)), val.(*Car).GetName())
					fmt.Printf("%v (returned object): %#v\n", reflect.TypeOf(val.(*Car).ItemArr[0]), val.(*Car).ItemArr[0].GetName())
				}
			case *[]Car:
				// use the original object.
				if name := (*test.obj.(*[]Car))[0].GetName(); name != test.want {
					t.Errorf("%v.GetName() = %v; want = %v", reflect.TypeOf(*test.obj.(*[]Car)), (*test.obj.(*[]Car))[0].GetName(), test.want)
				} else if name := (*test.obj.(*[]Car))[0].ItemArr[0].GetName(); name != test.want2 {
					t.Errorf("%v.GetName() = %v; want = %v", reflect.TypeOf((*test.obj.(*[]Car))[0].ItemArr[0]), (*test.obj.(*[]Car))[0].ItemArr[0].GetName(), test.want2)
				} else {
					fmt.Printf("%v (original object): %#v\n", reflect.TypeOf(*test.obj.(*[]Car)), (*test.obj.(*[]Car))[0].GetName())
					fmt.Printf("%v (original object): %#v\n", reflect.TypeOf((*test.obj.(*[]Car))[0].ItemArr[0]), (*test.obj.(*[]Car))[0].ItemArr[0].GetName())
				}

				// use the returned object.
				if name := (*val.(*[]Car))[0].GetName(); name != test.want {
					t.Errorf("%v.GetName() = %v; want = %v", reflect.TypeOf(*val.(*[]Car)), (*val.(*[]Car))[0].GetName(), test.want)
				} else if name := (*val.(*[]Car))[0].ItemArr[0].GetName(); name != test.want2 {
					t.Errorf("%v.GetName() = %v; want = %v", reflect.TypeOf((*val.(*[]Car))[0].ItemArr[0]), (*val.(*[]Car))[0].ItemArr[0].GetName(), test.want2)
				} else {
					fmt.Printf("%v (returned object): %#v\n", reflect.TypeOf(*val.(*[]Car)), (*val.(*[]Car))[0].GetName())
					fmt.Printf("%v (returned object): %#v\n", reflect.TypeOf((*val.(*[]Car))[0].ItemArr[0]), (*val.(*[]Car))[0].ItemArr[0].GetName())
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

	if err := o.AddData("test", "test"); err != nil {
		// do something
	}

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

		if err := o.AddData("Car", car); err != nil {
		}

		if err := o.AddData("Hello", "World"); err != nil {
		}

		if err := o.AddData("Age", 18); err != nil {
		}

		o.Encode()
	}
}
