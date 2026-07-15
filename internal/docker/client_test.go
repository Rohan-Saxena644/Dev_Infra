package docker

import (
	"encoding/json"
	"testing"
)

func TestSelectExposedPort(t *testing.T) {
	tests := []struct {
		name  string
		ports map[string]json.RawMessage
		want  int
	}{
		{name: "defaults to port 80", want: 80},
		{name: "uses Next.js port", ports: map[string]json.RawMessage{"3000/tcp": nil}, want: 3000},
		{name: "prefers port 80", ports: map[string]json.RawMessage{"3000/tcp": nil, "80/tcp": nil}, want: 80},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := selectExposedPort(test.ports)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("got %d, want %d", got, test.want)
			}
		})
	}
}

func TestSelectExposedPortRejectsAmbiguousImage(t *testing.T) {
	ports := map[string]json.RawMessage{"3000/tcp": nil, "8080/tcp": nil}
	if _, err := selectExposedPort(ports); err == nil {
		t.Fatal("expected ambiguous ports to fail")
	}
}
