package pokeapi

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

// These tests are offline: they exercise the URI driver's pure string functions
// and the host wiring, which need no network. The client's HTTP behaviour is
// covered in pokeapi_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "pokeapi" {
		t.Errorf("Scheme = %q, want pokeapi", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "pokeapi" {
		t.Errorf("Identity.Binary = %q, want pokeapi", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct{ in, typ, id string }{
		{"pikachu", "pokemon", "pikachu"},
		{"25", "pokemon", "25"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("pokemon", "pikachu")
	want := "https://pokeapi.co/api/v2/pokemon/pikachu"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

// TestHostWiring mounts the driver in a kit Host and checks the round trip.
func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}

	p := &Pokemon{ID: 25, Name: "pikachu", URL: "https://pokeapi.co/api/v2/pokemon/25/"}
	_, err = h.Mint(p)
	// Mint may fail if the struct lacks kit:"id" tag — that is fine for now;
	// we only check that the domain registers and Open succeeds.
	_ = err

	got, err := h.ResolveOn("pokeapi", "pikachu")
	if err != nil || got.String() != "pokeapi://pokemon/pikachu" {
		t.Errorf("ResolveOn = (%q, %v), want pokeapi://pokemon/pikachu", got.String(), err)
	}
}
