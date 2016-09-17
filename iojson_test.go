package iojson

import (
	"fmt"
	//"reflect"
	"strings"
	"testing"
)

type Car struct {
	Name string
}

func (c *Car) GetName() string {
	return c.Name
}

func TestGetData(t *testing.T) {
	var tests = []struct {
		json        string
		key         string
		keyNotExist string
		want        string
		obj         interface{}
	}{
		{`{"Data":{"%s":{"Name":"%s"}}}`, "Car", "", "BMW", &Car{Name: "Init Car"}},
		{`{"Data":{"%s":"%s"}}`, "Hello", "", "World", nil},
		{`{"Data":{"%s":%s}}`, "Amt", "", "123.8", nil},
		{`{"Data":{"%s":%s}}`, "Amt", "X", "123.8", nil},
	}

	for _, test := range tests {
		//fmt.Printf("HERE: %v\n", reflect.TypeOf(test.obj))
		//theType := reflect.TypeOf(test.obj)
		//theType := reflect.New(reflect.TypeOf(test.obj)).Interface()

		test.json = fmt.Sprintf(test.json, test.key, test.want)
		//fmt.Printf("HERE: %v\n", test.json)

		test.key += test.keyNotExist

		i := NewIOJSON()

		if err := i.Decode(strings.NewReader(test.json)); err != nil {
			t.Errorf("i.Decode(strings.NewReader(%v)) = %v", test.json, err)

			continue
		} else if val, err := i.GetData(test.key, test.obj); err != nil {
			if err.Error() == test.key+ErrDataKeyNotExist {
				// Do nothing. Recognized error.
			} else {
				t.Errorf("i.GetData(%v) = %v", test.key, err)
			}

			continue
		} else if test.obj == nil {
			fmt.Printf("%v: %#v\n", test.key, val)
		}

		if test.obj != nil {
			switch v := test.obj.(type) {
			case *Car:
				if name := test.obj.(*Car).GetName(); name != test.want {
					t.Errorf("%v.GetName(%v) = %v", test.key, test.key, name)
				} else {
					fmt.Printf("%v: %#v\n", test.key, name)
				}
			default:
				t.Errorf("test.obj(type) = %v", v)
			}
		}
	}

	//carName := "BMW"
	//str := fmt.Sprintf(`{"Data":{"Hello":"World","Car":{"Name":"%s"}, "Age":18 }}`, carName)

	//i := NewIOJSON()
	//i.Decode(strings.NewReader(str))

	//if _, err := i.GetData("Car", car); err != nil {
	//	t.Errorf("i.GetData(%v) = %v", "Car", err)
	//} else if n := car.GetName(); n != carName {
	//	t.Errorf("car.GetName(%v) = %v", carName, n)
	//}
}

func BenchmarkEncode(b *testing.B) {
	o := NewIOJSON()

	o.AddData("test", "test")

	for i := 0; i < b.N; i++ {
		o.Encode()
	}
}
