package main

import "testing"

func FailIfFalse(value bool, message string, t *testing.T) {
	if !value {
		t.Error(message)
	}
}

func FailNowIfFalse(value bool, message string, t *testing.T) {
	if !value {
		t.Fatal(message)
	}
}
