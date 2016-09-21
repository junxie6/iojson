package iojson

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

import (
	"github.com/junhsieh/util"
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
	want        string // Car value
	want2       string // Item value
	obj         interface{}
}

func GetTestCase(storeType string) []TestCase {
	var TestCaseMap = map[string][]TestCase{}

	TestCaseMap["ObjArr"] = []TestCase{
		// {[]}
		{`{"` + storeType + `":[{"Name":"%s","ItemArr":[{"Name":"%s"}]}]}`, "0", "", "My luxury car", "Bag", &Car{}},
		// []{[]}
		{`{"` + storeType + `":[[{"Name":"%s","ItemArr":[{"Name":"%s"}]}]]}`, "0", "", "My luxury car", "Bag", &[]Car{}},
		// ""
		{`{"` + storeType + `":["%s"]}`, "0", "", "World", "", nil},
		// 0
		{`{"` + storeType + `":[%s]}`, "0", "", "123.8", "", nil},
		// null
		{`{"` + storeType + `":[%s]}`, "0", "", "null", "", nil},
		// {}
		{`{"` + storeType + `":[%s]}`, "0", "", "{}", "", nil},
		// ""
		{`{"` + storeType + `":["%s"]}`, "0", "", "", "", nil},
		//
		{`{"` + storeType + `":[%s]}`, "0", "", "", "", nil},
	}

	TestCaseMap["ObjMap"] = []TestCase{
		// {[]}
		{`{"` + storeType + `":{"%s":{"Name":"%s","ItemArr":[{"Name":"%s"}]}}}`, "Car", "", "My luxury car", "Bag", &Car{}},
		{`{"` + storeType + `":{"%s":{"Name":"%s","ItemArr":[{"Name":"%s"}]}}}`, "Car", "Dummy", "My luxury car", "Bag", &Car{}},
		// []{[]}
		{`{"` + storeType + `":{"%s":[{"Name":"%s","ItemArr":[{"Name":"%s"}]}]}}`, "Car", "", "My luxury car", "Bag", &[]Car{}},
		{`{"` + storeType + `":{"%s":[{"Name":"%s","ItemArr":[{"Name":"%s"}]}]}}`, "Car", "Dummy", "My luxury car", "Bag", &[]Car{}},
		// ""
		{`{"` + storeType + `":{"%s":"%s"}}`, "Hello", "", "World", "", nil},
		{`{"` + storeType + `":{"%s":"%s"}}`, "Hello", "Dummy", "World", "", nil},
		// 0
		{`{"` + storeType + `":{"%s":%s}}`, "Amt", "", "123.8", "", nil},
		{`{"` + storeType + `":{"%s":%s}}`, "Amt", "Dummy", "123.8", "", nil},
		// null
		{`{"` + storeType + `":{"%s":%s}}`, "Null", "", "null", "", nil},
		{`{"` + storeType + `":{"%s":%s}}`, "Null", "Dummy", "null", "", nil},
		// {}
		{`{"` + storeType + `":{"%s":%s}}`, "Braces", "", "{}", "", nil},
		{`{"` + storeType + `":{"%s":%s}}`, "Braces", "Dummy", "{}", "", nil},
		// ""
		{`{"` + storeType + `":{"%s":"%s"}}`, "Empty", "", "", "", nil},
		{`{"` + storeType + `":{"%s":"%s"}}`, "Empty", "Dummy", "", "", nil},
		//
		{`{"` + storeType + `":{}}`, "None", "", "", "", nil},
		{`{"` + storeType + `":{}}`, "None", "Dummy", "", "", nil},
	}

	var v []TestCase
	var ok bool

	if v, ok = TestCaseMap[storeType]; !ok {
		return nil
	}

	return v
}

func TestAddObjToArr(t *testing.T) {
	s1 := `{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[{"Name":"My luxury car","ItemArr":null},{"Name":"Bag","ItemArr":null}],"ObjMap":{}}`

	car := &Car{Name: "My luxury car"}
	item := &Car{Name: "Bag"}

	o := NewIOJSON()

	o.AddObjToArr(car)
	o.AddObjToArr(item)

	if ok, err := util.JSONDeepEqual(o.EncodeString(), s1); err != nil {
		t.Errorf("err: %v", err)
	} else if !ok {
		t.Errorf("util.JSONDeepEqual(%s, %s) = %v", o.EncodeString(), s1, ok)
	}
}

func TestAddObjToMap(t *testing.T) {
	s1 := `{"Status":true,"ErrArr":[],"ErrCount":0,"ObjArr":[],"ObjMap":{"Car":{"Name":"My luxury car","ItemArr":null},"Item":{"Name":"Bag","ItemArr":null}}}`

	car := &Car{Name: "My luxury car"}
	item := &Car{Name: "Bag"}

	o := NewIOJSON()

	o.AddObjToMap("Car", car)
	o.AddObjToMap("Item", item)

	if ok, err := util.JSONDeepEqual(o.EncodeString(), s1); err != nil {
		t.Errorf("err: %v", err)
	} else if !ok {
		t.Errorf("util.JSONDeepEqual(%s, %s) = %v", o.EncodeString(), s1, ok)
	}
}

func TestGetObj(t *testing.T) {
	testTypeArr := []string{
		"ObjArr",
		"ObjMap",
	}

	for _, testtype := range testTypeArr {
		fmt.Printf("================= [ %s ] ================\n", testtype)

		var tests = GetTestCase(testtype)

		for _, test := range tests {
			//fmt.Printf("HERE: %v\n", reflect.TypeOf(test.obj))
			//theType := reflect.New(reflect.TypeOf(test.obj)).Interface()

			switch testtype {
			case "ObjArr":
				test.json = fmt.Sprintf(test.json, test.want, test.want2)
			case "ObjMap":
				test.json = fmt.Sprintf(test.json, test.key, test.want, test.want2)
			}
			//fmt.Printf("HERE: %v\n", test.json)

			test.key += test.keyNotExist

			i := NewIOJSON()

			if err := i.Decode(strings.NewReader(test.json)); err != nil {
				t.Errorf("i.Decode(strings.NewReader(%v)) = %v", test.json, err)

				continue
			}

			var val interface{}
			var err error

			switch testtype {
			case "ObjArr":
				index, _ := strconv.Atoi(test.key)
				val, err = i.GetObjFromArr(index, test.obj)
			case "ObjMap":
				val, err = i.GetObjFromMap(test.key, test.obj)
			}

			if err != nil {
				if err.Error() == test.key+ErrKeyNotExist {
					// Do nothing. Recognized error.
					fmt.Printf("%v (not exist): %#v\n", test.key, val)
				} else if err.Error() == ErrJSONRawIsNil && test.want == "null" {
					// Do nothing. Recognized error.
					fmt.Printf("%v (null): %#v\n", test.key, val)
				} else {
					t.Errorf("i.Get"+testtype+"(%v, %v) = %v", test.key, test.obj, err)
				}

				continue
			}

			//continue

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

	fmt.Printf("================= [ %s ] ================\n", "END")
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		car := &Car{
			Name: "Init Car",
		}

		o := NewIOJSON()

		if err := o.AddObjToMap("Car", car); err != nil {
		}

		if err := o.AddObjToMap("Hello", "World"); err != nil {
		}

		if err := o.AddObjToMap("Age", 18); err != nil {
		}

		o.Encode()
	}
}
