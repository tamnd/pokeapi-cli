package pokeapi_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/pokeapi-cli/pokeapi"
)

// --- mock JSON payloads ---

const fakeListJSON = `{
  "count": 1350,
  "next": "https://pokeapi.co/api/v2/pokemon?offset=3&limit=3",
  "previous": null,
  "results": [
    {"name": "bulbasaur", "url": "https://pokeapi.co/api/v2/pokemon/1/"},
    {"name": "ivysaur",   "url": "https://pokeapi.co/api/v2/pokemon/2/"},
    {"name": "venusaur",  "url": "https://pokeapi.co/api/v2/pokemon/3/"}
  ]
}`

const fakePikachuJSON = `{
  "id": 25,
  "name": "pikachu",
  "base_experience": 112,
  "height": 4,
  "weight": 60,
  "types": [
    {"slot": 1, "type": {"name": "electric", "url": "https://pokeapi.co/api/v2/type/13/"}}
  ],
  "abilities": [
    {"ability": {"name": "static", "url": "..."}, "is_hidden": false},
    {"ability": {"name": "lightning-rod", "url": "..."}, "is_hidden": true}
  ],
  "stats": [
    {"base_stat": 35, "effort": 0, "stat": {"name": "hp",              "url": "..."}},
    {"base_stat": 55, "effort": 0, "stat": {"name": "attack",          "url": "..."}},
    {"base_stat": 40, "effort": 0, "stat": {"name": "defense",         "url": "..."}},
    {"base_stat": 50, "effort": 0, "stat": {"name": "special-attack",  "url": "..."}},
    {"base_stat": 50, "effort": 0, "stat": {"name": "special-defense", "url": "..."}},
    {"base_stat": 90, "effort": 2, "stat": {"name": "speed",           "url": "..."}}
  ],
  "sprites": {
    "front_default": "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/25.png",
    "back_default": "...",
    "front_shiny": "..."
  },
  "species": {"name": "pikachu", "url": "https://pokeapi.co/api/v2/pokemon-species/25/"},
  "abilities": [],
  "moves": []
}`

const fakeSpeciesJSON = `{
  "id": 25,
  "name": "pikachu",
  "is_legendary": false,
  "is_mythical": false,
  "capture_rate": 190,
  "base_happiness": 70,
  "color": {"name": "yellow", "url": "..."},
  "generation": {"name": "generation-i", "url": "..."},
  "flavor_text_entries": [
    {
      "flavor_text": "When several of these POKéMON gather,\ntheir electricity can cause lightning storms.",
      "language": {"name": "en", "url": "..."},
      "version": {"name": "red", "url": "..."}
    }
  ]
}`

const fakeTypeFireJSON = `{
  "id": 10,
  "name": "fire",
  "pokemon": [
    {"pokemon": {"name": "charmander", "url": "..."}, "slot": 1},
    {"pokemon": {"name": "charmeleon", "url": "..."}, "slot": 1},
    {"pokemon": {"name": "charizard",  "url": "..."}, "slot": 1}
  ],
  "damage_relations": {
    "double_damage_to": [
      {"name": "bug",   "url": "..."},
      {"name": "steel", "url": "..."},
      {"name": "grass", "url": "..."},
      {"name": "ice",   "url": "..."}
    ],
    "double_damage_from": [
      {"name": "water",  "url": "..."},
      {"name": "ground", "url": "..."},
      {"name": "rock",   "url": "..."}
    ],
    "half_damage_to": [
      {"name": "fire",   "url": "..."},
      {"name": "water",  "url": "..."},
      {"name": "rock",   "url": "..."},
      {"name": "dragon", "url": "..."}
    ],
    "no_damage_to": []
  }
}`

const fakeAbilityJSON = `{
  "id": 65,
  "name": "overgrow",
  "pokemon": [
    {"pokemon": {"name": "bulbasaur", "url": "..."}, "is_hidden": false},
    {"pokemon": {"name": "ivysaur",   "url": "..."}, "is_hidden": false}
  ],
  "effect_entries": [
    {
      "effect": "When this Pokémon has 1/3 or less of its HP remaining, its grass-type moves inflict 1.5× as much regular damage.",
      "short_effect": "Strengthens grass moves to inflict 1.5× damage at 1/3 max HP or less.",
      "language": {"name": "en", "url": "..."}
    }
  ]
}`

const fakeMoveJSON = `{
  "id": 53,
  "name": "flamethrower",
  "power": 90,
  "accuracy": 100,
  "pp": 15,
  "type": {"name": "fire", "url": "..."},
  "damage_class": {"name": "special", "url": "..."},
  "effect_entries": [
    {
      "effect": "Inflicts regular damage. Has a $effect_chance% chance to burn the target.",
      "language": {"name": "en", "url": "..."}
    }
  ]
}`

const fakeGenerationJSON = `{
  "id": 1,
  "name": "generation-i",
  "main_region": {"name": "kanto", "url": "..."},
  "pokemon_species": [
    {"name": "bulbasaur", "url": "..."},
    {"name": "ivysaur",   "url": "..."},
    {"name": "venusaur",  "url": "..."}
  ],
  "types": [
    {"name": "normal",  "url": "..."},
    {"name": "fighting","url": "..."}
  ]
}`

const fakeItemJSON = `{
  "id": 17,
  "name": "potion",
  "cost": 300,
  "fling_power": 30,
  "category": {"name": "healing", "url": "..."},
  "effect_entries": [
    {
      "effect": "Restores 20 HP.",
      "language": {"name": "en", "url": "..."}
    }
  ]
}`

const fakeBerryJSON = `{
  "id": 1,
  "name": "cheri",
  "growth_time": 3,
  "max_harvest": 5,
  "natural_gift_power": 60,
  "size": 20,
  "smoothness": 25,
  "soil_dryness": 15,
  "firmness": {"name": "soft", "url": "..."},
  "flavors": [{"potency": 0, "flavor": {"name": "spicy", "url": "..."}}],
  "natural_gift_type": {"name": "fire", "url": "..."}
}`

// --- helpers ---

func newTestClient(ts *httptest.Server) *pokeapi.Client {
	cfg := pokeapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return pokeapi.NewClient(cfg)
}

func serve(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, body)
	}))
}

// --- list tests ---

func TestListSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeListJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.List(context.Background(), 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestListParsesItems(t *testing.T) {
	ts := serve(fakeListJSON)
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.List(context.Background(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}
	if items[0].Rank != 1 {
		t.Errorf("items[0].Rank = %d, want 1", items[0].Rank)
	}
	if items[0].Name != "bulbasaur" {
		t.Errorf("items[0].Name = %q, want bulbasaur", items[0].Name)
	}
	if items[1].Name != "ivysaur" {
		t.Errorf("items[1].Name = %q, want ivysaur", items[1].Name)
	}
	if items[2].Name != "venusaur" {
		t.Errorf("items[2].Name = %q, want venusaur", items[2].Name)
	}
}

func TestListLimitRespected(t *testing.T) {
	ts := serve(fakeListJSON)
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.List(context.Background(), 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d, want 2", len(items))
	}
}

func TestListRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeListJSON)
	}))
	defer ts.Close()

	cfg := pokeapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := pokeapi.NewClient(cfg)

	_, err := c.List(context.Background(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

// --- pokemon (get) tests ---

func TestGetPokemonParses(t *testing.T) {
	ts := serve(fakePikachuJSON)
	defer ts.Close()

	c := newTestClient(ts)
	p, err := c.GetPokemon(context.Background(), "pikachu")
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != 25 {
		t.Errorf("p.ID = %d, want 25", p.ID)
	}
	if p.Name != "pikachu" {
		t.Errorf("p.Name = %q, want pikachu", p.Name)
	}
	if len(p.Types) != 1 || p.Types[0] != "electric" {
		t.Errorf("p.Types = %v, want [electric]", p.Types)
	}
	if p.HP != 35 {
		t.Errorf("p.HP = %d, want 35", p.HP)
	}
	if p.Attack != 55 {
		t.Errorf("p.Attack = %d, want 55", p.Attack)
	}
	if p.Defense != 40 {
		t.Errorf("p.Defense = %d, want 40", p.Defense)
	}
	if p.SpAttack != 50 {
		t.Errorf("p.SpAttack = %d, want 50", p.SpAttack)
	}
	if p.SpDefense != 50 {
		t.Errorf("p.SpDefense = %d, want 50", p.SpDefense)
	}
	if p.Speed != 90 {
		t.Errorf("p.Speed = %d, want 90", p.Speed)
	}
	if !strings.Contains(p.SpriteURL, "25.png") {
		t.Errorf("p.SpriteURL = %q, want to contain 25.png", p.SpriteURL)
	}
}

// --- species tests ---

func TestGetSpeciesParses(t *testing.T) {
	ts := serve(fakeSpeciesJSON)
	defer ts.Close()

	c := newTestClient(ts)
	s, err := c.GetSpecies(context.Background(), "pikachu")
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != 25 {
		t.Errorf("s.ID = %d, want 25", s.ID)
	}
	if s.IsLegendary {
		t.Error("s.IsLegendary = true, want false")
	}
	if s.IsMythical {
		t.Error("s.IsMythical = true, want false")
	}
	if s.Color != "yellow" {
		t.Errorf("s.Color = %q, want yellow", s.Color)
	}
	if s.Generation != "generation-i" {
		t.Errorf("s.Generation = %q, want generation-i", s.Generation)
	}
	if s.CaptureRate != 190 {
		t.Errorf("s.CaptureRate = %d, want 190", s.CaptureRate)
	}
	if s.FlavorText == "" {
		t.Error("s.FlavorText is empty")
	}
	// newlines should be replaced with spaces
	if strings.Contains(s.FlavorText, "\n") {
		t.Errorf("s.FlavorText contains newline: %q", s.FlavorText)
	}
}

// --- type tests ---

func TestGetTypeParses(t *testing.T) {
	ts := serve(fakeTypeFireJSON)
	defer ts.Close()

	c := newTestClient(ts)
	typ, err := c.GetType(context.Background(), "fire")
	if err != nil {
		t.Fatal(err)
	}
	if typ.Name != "fire" {
		t.Errorf("typ.Name = %q, want fire", typ.Name)
	}
	if typ.PokemonCount != 3 {
		t.Errorf("typ.PokemonCount = %d, want 3", typ.PokemonCount)
	}
	wantSE := []string{"bug", "steel", "grass", "ice"}
	for _, w := range wantSE {
		found := false
		for _, got := range typ.SuperEffective {
			if got == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("typ.SuperEffective missing %q, got %v", w, typ.SuperEffective)
		}
	}
	if len(typ.WeakTo) != 3 {
		t.Errorf("typ.WeakTo = %v, want 3 entries", typ.WeakTo)
	}
	if len(typ.NoEffect) != 0 {
		t.Errorf("typ.NoEffect = %v, want empty", typ.NoEffect)
	}
}

// --- ability tests ---

func TestGetAbilityParses(t *testing.T) {
	ts := serve(fakeAbilityJSON)
	defer ts.Close()

	c := newTestClient(ts)
	a, err := c.GetAbility(context.Background(), "overgrow")
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "overgrow" {
		t.Errorf("a.Name = %q, want overgrow", a.Name)
	}
	if a.PokemonCount != 2 {
		t.Errorf("a.PokemonCount = %d, want 2", a.PokemonCount)
	}
	if a.Effect == "" {
		t.Error("a.Effect is empty")
	}
	if a.ShortEffect == "" {
		t.Error("a.ShortEffect is empty")
	}
}

// --- move tests ---

func TestGetMoveParses(t *testing.T) {
	ts := serve(fakeMoveJSON)
	defer ts.Close()

	c := newTestClient(ts)
	m, err := c.GetMove(context.Background(), "flamethrower")
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "flamethrower" {
		t.Errorf("m.Name = %q, want flamethrower", m.Name)
	}
	if m.Power != 90 {
		t.Errorf("m.Power = %d, want 90", m.Power)
	}
	if m.Accuracy != 100 {
		t.Errorf("m.Accuracy = %d, want 100", m.Accuracy)
	}
	if m.PP != 15 {
		t.Errorf("m.PP = %d, want 15", m.PP)
	}
	if m.Type != "fire" {
		t.Errorf("m.Type = %q, want fire", m.Type)
	}
	if m.DamageClass != "special" {
		t.Errorf("m.DamageClass = %q, want special", m.DamageClass)
	}
	if m.Effect == "" {
		t.Error("m.Effect is empty")
	}
}

// --- generation tests ---

func TestGetGenerationParses(t *testing.T) {
	ts := serve(fakeGenerationJSON)
	defer ts.Close()

	c := newTestClient(ts)
	g, err := c.GetGeneration(context.Background(), "generation-i")
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "generation-i" {
		t.Errorf("g.Name = %q, want generation-i", g.Name)
	}
	if g.MainRegion != "kanto" {
		t.Errorf("g.MainRegion = %q, want kanto", g.MainRegion)
	}
	if g.PokemonCount != 3 {
		t.Errorf("g.PokemonCount = %d, want 3", g.PokemonCount)
	}
	if g.TypeCount != 2 {
		t.Errorf("g.TypeCount = %d, want 2", g.TypeCount)
	}
}

// --- berry tests ---

func TestGetBerryParses(t *testing.T) {
	ts := serve(fakeBerryJSON)
	defer ts.Close()

	c := newTestClient(ts)
	b, err := c.GetBerry(context.Background(), "cheri")
	if err != nil {
		t.Fatal(err)
	}
	if b.ID != 1 {
		t.Errorf("b.ID = %d, want 1", b.ID)
	}
	if b.Name != "cheri" {
		t.Errorf("b.Name = %q, want cheri", b.Name)
	}
	if b.GrowthTime != 3 {
		t.Errorf("b.GrowthTime = %d, want 3", b.GrowthTime)
	}
	if b.MaxHarvest != 5 {
		t.Errorf("b.MaxHarvest = %d, want 5", b.MaxHarvest)
	}
	if b.Firmness != "soft" {
		t.Errorf("b.Firmness = %q, want soft", b.Firmness)
	}
	if b.GiftType != "fire" {
		t.Errorf("b.GiftType = %q, want fire", b.GiftType)
	}
	if b.GiftPower != 60 {
		t.Errorf("b.GiftPower = %d, want 60", b.GiftPower)
	}
}

// --- item tests ---

func TestGetItemParses(t *testing.T) {
	ts := serve(fakeItemJSON)
	defer ts.Close()

	c := newTestClient(ts)
	it, err := c.GetItem(context.Background(), "potion")
	if err != nil {
		t.Fatal(err)
	}
	if it.Name != "potion" {
		t.Errorf("it.Name = %q, want potion", it.Name)
	}
	if it.Cost != 300 {
		t.Errorf("it.Cost = %d, want 300", it.Cost)
	}
	if it.Category != "healing" {
		t.Errorf("it.Category = %q, want healing", it.Category)
	}
	if it.Effect == "" {
		t.Error("it.Effect is empty")
	}
}
