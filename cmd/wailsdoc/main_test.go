package main

import "testing"

func TestExplicitVersion(t *testing.T) {
	version = "v1.2.3"
	t.Cleanup(func() { version = "" })
	if actual := buildVersion(); actual != version {
		t.Fatalf("buildVersion() = %q", actual)
	}
}
