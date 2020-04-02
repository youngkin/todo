package todo

import "testing"

func TestValidate(t *testing.T) {
	item := Item{
		Note: "",
	}

	err := validateToDo(item)
	if err == nil {
		t.Error("expected error")
	}
}
