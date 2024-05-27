package JSON

import (
	"testing"
)

func TestStringify(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	person := Person{"John", 30}
	expected := `{"Name":"John","Age":30}`
	result := Stringify(person)
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestParse(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	jsonString := `{"Name":"John","Age":30}`
	expected := Person{"John", 30}
	var result Person
	err := Parse(jsonString, &result)
	if err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}
