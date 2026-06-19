package main

import (
	"reflect"
	"testing"
)

func TestNormalizeJoinArgsAcceptsPasswordAfterCode(t *testing.T) {
	got := normalizeJoinArgs([]string{"7906789", "--password", "secret"})
	want := []string{"--password", "secret", "7906789"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeJoinArgs() = %#v, want %#v", got, want)
	}
}

func TestNormalizeJoinArgsAcceptsPasswordBeforeCode(t *testing.T) {
	got := normalizeJoinArgs([]string{"--password", "secret", "7906789"})
	want := []string{"--password", "secret", "7906789"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeJoinArgs() = %#v, want %#v", got, want)
	}
}

func TestNormalizeJoinArgsAcceptsBarePasswordToken(t *testing.T) {
	got := normalizeJoinArgs([]string{"7906789", "password", "secret"})
	want := []string{"--password", "secret", "7906789"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeJoinArgs() = %#v, want %#v", got, want)
	}
}
