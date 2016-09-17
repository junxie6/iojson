package iojson

import (
	"fmt"
	"strings"
	//"reflect"
	"testing"
)

type Car struct {
	Name string
}

func (c *Car) GetName() string {
	return c.Name
}

func TestGetData(t *testing.T) {
	carName := "BMW"
	str := fmt.Sprintf(`{"Data":{"Hello":"World","Car":{"Name":"%s"}, "Age":18 }}`, carName)

	i := NewIOJSON()
	i.Decode(strings.NewReader(str))

	car := &Car{
		Name: "Init Car",
	}

	if _, err := i.GetData("Car", car); err != nil {
		t.Errorf("i.GetData(%v) = %v", "Car", err)
	} else if n := car.GetName(); n != carName {
		t.Errorf("car.GetName(%v) = %v", carName, n)
	}
}

func BenchmarkEncode(b *testing.B) {
	o := NewIOJSON()

	o.AddData("test", "test")

	for i := 0; i < b.N; i++ {
		o.Encode()
	}
}
