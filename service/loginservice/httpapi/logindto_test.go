package httpapi

import (
	"reflect"
	"testing"
)

func TestLoginRequestOnlyContainsAuthenticationFields(t *testing.T) {
	typeOfRequest := reflect.TypeOf(LoginRequest{})
	if typeOfRequest.NumField() != 3 {
		t.Fatalf("LoginRequest field count = %d, want 3", typeOfRequest.NumField())
	}
	for _, name := range []string{"PlatType", "PlatID", "AccessToken"} {
		if _, exists := typeOfRequest.FieldByName(name); !exists {
			t.Fatalf("LoginRequest missing field %s", name)
		}
	}
}
