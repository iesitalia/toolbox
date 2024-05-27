package toolbox

import (
	"testing"
)

func TestDictionary(t *testing.T) {
	// Test Has method
	t.Run("Has", func(t *testing.T) {
		dict := Dictionary[string]{
			{Key: "name", Value: "Alice"},
			{Key: "city", Value: "Wonderland"},
		}

		if !dict.Has("name") {
			t.Error("Expected to have key 'name'")
		}

		if dict.Has("age") {
			t.Error("Expected not to have key 'age'")
		}
	})

	// Test ContainsValue method
	t.Run("ContainsValue", func(t *testing.T) {
		dict := Dictionary[string]{
			{Key: "name", Value: "Alice"},
			{Key: "city", Value: "Wonderland"},
		}

		found, kv := dict.ContainsValue("Alice")
		if !found || kv.Key != "name" {
			t.Error("Expected to contain value 'Alice' with key 'name'")
		}

		found, _ = dict.ContainsValue("Bob")
		if found {
			t.Error("Expected not to contain value 'Bob'")
		}
	})

	// Test Set method
	t.Run("Set", func(t *testing.T) {
		dict := Dictionary[string]{}
		dict.Set("name", "Alice")

		if len(dict) != 1 {
			t.Errorf("Expected dictionary length 1, got %d", len(dict))
		}

		if dict[0].Key != "name" || dict[0].Value != "Alice" {
			t.Error("Expected key-value pair 'name'-'Alice'")
		}

		dict.Set("name", "Bob")
		if dict[0].Value != "Bob" {
			t.Error("Expected key 'name' to be updated to value 'Bob'")
		}
	})

	// Test Delete method
	t.Run("Delete", func(t *testing.T) {
		dict := Dictionary[string]{
			{Key: "name", Value: "Alice"},
			{Key: "city", Value: "Wonderland"},
		}

		err := dict.Delete("name")
		if err != nil {
			t.Error("Expected to delete key 'name'")
		}

		if len(dict) != 1 {
			t.Errorf("Expected dictionary length 1, got %d", len(dict))
		}

		if dict.Has("name") {
			t.Error("Expected key 'name' to be deleted")
		}

		err = dict.Delete("non-existing")
		if err == nil {
			t.Error("Expected error when deleting non-existing key")
		}
	})

	// Test ContainsValue with non-string type
	t.Run("ContainsValue with int", func(t *testing.T) {
		dict := Dictionary[int]{
			{Key: "age", Value: 30},
			{Key: "score", Value: 100},
		}

		found, kv := dict.ContainsValue(30)
		if !found || kv.Key != "age" {
			t.Error("Expected to contain value 30 with key 'age'")
		}

		found, _ = dict.ContainsValue(50)
		if found {
			t.Error("Expected not to contain value 50")
		}
	})

	// Test Set with non-string type
	t.Run("Set with int", func(t *testing.T) {
		dict := Dictionary[int]{}
		dict.Set("age", 30)

		if len(dict) != 1 {
			t.Errorf("Expected dictionary length 1, got %d", len(dict))
		}

		if dict[0].Key != "age" || dict[0].Value != 30 {
			t.Error("Expected key-value pair 'age'-30")
		}

		dict.Set("age", 40)
		if dict[0].Value != 40 {
			t.Error("Expected key 'age' to be updated to value 40")
		}
	})

	// Test Delete with non-string type
	t.Run("Delete with int", func(t *testing.T) {
		dict := Dictionary[int]{
			{Key: "age", Value: 30},
			{Key: "score", Value: 100},
		}

		err := dict.Delete("age")
		if err != nil {
			t.Error("Expected to delete key 'age'")
		}

		if len(dict) != 1 {
			t.Errorf("Expected dictionary length 1, got %d", len(dict))
		}

		if dict.Has("age") {
			t.Error("Expected key 'age' to be deleted")
		}

		err = dict.Delete("non-existing")
		if err == nil {
			t.Error("Expected error when deleting non-existing key")
		}
	})
}
