// Package pokeapi is the library behind the pokeapi command line:
// the HTTP client, request shaping, and the typed data models for PokéAPI.
//
// PokéAPI is free and open, no auth required. The client sets a real
// User-Agent, paces requests to stay polite, and retries transient failures
// (429 and 5xx) with exponential backoff.
package pokeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "pokeapi.co"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://pokeapi.co/api/v2",
		UserAgent: "pokeapi-cli/0.1.0 (github.com/tamnd/pokeapi-cli)",
		Rate:      300 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
	}
}

// Client talks to PokéAPI over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// List returns PokemonListItems starting at offset. Rank is 1-based (offset+index+1).
// Pass limit=0 for API default of 20.
func (c *Client) List(ctx context.Context, limit, offset int) ([]PokemonListItem, error) {
	n := limit
	if n <= 0 {
		n = 20
	}
	u := fmt.Sprintf("%s/pokemon?limit=%d&offset=%d", c.cfg.BaseURL, n, offset)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp listResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode list: %w", err)
	}
	items := make([]PokemonListItem, 0, len(resp.Results))
	for i, r := range resp.Results {
		items = append(items, PokemonListItem{
			Rank: offset + i + 1,
			Name: r.Name,
			URL:  r.URL,
		})
	}
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items, nil
}

// GetPokemon fetches the full detail record for the named or numbered Pokémon.
func (c *Client) GetPokemon(ctx context.Context, nameOrID string) (*Pokemon, error) {
	u := fmt.Sprintf("%s/pokemon/%s", c.cfg.BaseURL, nameOrID)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw pokemonResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode pokemon: %w", err)
	}

	types := make([]string, 0, len(raw.Types))
	for _, t := range raw.Types {
		types = append(types, t.Type.Name)
	}

	abilities := make([]string, 0, len(raw.Abilities))
	for _, a := range raw.Abilities {
		abilities = append(abilities, a.Ability.Name)
	}

	var hp, attack, defense, spAttack, spDefense, speed int
	for _, s := range raw.Stats {
		switch s.Stat.Name {
		case "hp":
			hp = s.BaseStat
		case "attack":
			attack = s.BaseStat
		case "defense":
			defense = s.BaseStat
		case "special-attack":
			spAttack = s.BaseStat
		case "special-defense":
			spDefense = s.BaseStat
		case "speed":
			speed = s.BaseStat
		}
	}

	return &Pokemon{
		ID:             raw.ID,
		Name:           raw.Name,
		Height:         raw.Height,
		Weight:         raw.Weight,
		BaseExperience: raw.BaseExperience,
		Types:          types,
		Abilities:      abilities,
		HP:             hp,
		Attack:         attack,
		Defense:        defense,
		SpAttack:       spAttack,
		SpDefense:      spDefense,
		Speed:          speed,
		SpriteURL:      raw.Sprites.FrontDefault,
		URL:            fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%d/", raw.ID),
	}, nil
}

// GetSpecies fetches the species record for the named or numbered Pokémon.
func (c *Client) GetSpecies(ctx context.Context, nameOrID string) (*Species, error) {
	u := fmt.Sprintf("%s/pokemon-species/%s", c.cfg.BaseURL, nameOrID)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw speciesResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode species: %w", err)
	}
	return &Species{
		ID:            raw.ID,
		Name:          raw.Name,
		IsLegendary:   raw.IsLegendary,
		IsMythical:    raw.IsMythical,
		CaptureRate:   raw.CaptureRate,
		BaseHappiness: raw.BaseHappiness,
		Color:         raw.Color.Name,
		Generation:    raw.Generation.Name,
		FlavorText:    firstEnglishFlavor(raw.FlavorTextEntries),
	}, nil
}

// GetType fetches the type record by name or id.
func (c *Client) GetType(ctx context.Context, nameOrID string) (*Type, error) {
	u := fmt.Sprintf("%s/type/%s", c.cfg.BaseURL, nameOrID)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw typeResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode type: %w", err)
	}
	return &Type{
		ID:               raw.ID,
		Name:             raw.Name,
		PokemonCount:     len(raw.Pokemon),
		SuperEffective:   nameURLSlice(raw.DamageRelations.DoubleDamageTo),
		WeakTo:           nameURLSlice(raw.DamageRelations.DoubleDamageFrom),
		NotVeryEffective: nameURLSlice(raw.DamageRelations.HalfDamageTo),
		NoEffect:         nameURLSlice(raw.DamageRelations.NoDamageTo),
	}, nil
}

// GetAbility fetches the ability record by name or id.
func (c *Client) GetAbility(ctx context.Context, nameOrID string) (*Ability, error) {
	u := fmt.Sprintf("%s/ability/%s", c.cfg.BaseURL, nameOrID)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw abilityResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode ability: %w", err)
	}
	effect, short := firstEnglishEffect(raw.EffectEntries)
	return &Ability{
		ID:           raw.ID,
		Name:         raw.Name,
		PokemonCount: len(raw.Pokemon),
		Effect:       effect,
		ShortEffect:  short,
	}, nil
}

// GetMove fetches the move record by name or id.
func (c *Client) GetMove(ctx context.Context, nameOrID string) (*Move, error) {
	u := fmt.Sprintf("%s/move/%s", c.cfg.BaseURL, nameOrID)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw moveResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode move: %w", err)
	}
	effect, _ := firstEnglishEffect(raw.EffectEntries)
	return &Move{
		ID:          raw.ID,
		Name:        raw.Name,
		Power:       raw.Power,
		Accuracy:    raw.Accuracy,
		PP:          raw.PP,
		Type:        raw.Type.Name,
		DamageClass: raw.DamageClass.Name,
		Effect:      effect,
	}, nil
}

// GetGeneration fetches the generation record by name or id.
func (c *Client) GetGeneration(ctx context.Context, nameOrID string) (*Generation, error) {
	u := fmt.Sprintf("%s/generation/%s", c.cfg.BaseURL, nameOrID)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw generationResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode generation: %w", err)
	}
	return &Generation{
		ID:           raw.ID,
		Name:         raw.Name,
		MainRegion:   raw.MainRegion.Name,
		PokemonCount: len(raw.PokemonSpecies),
		TypeCount:    len(raw.Types),
	}, nil
}

// GetBerry fetches the berry record by name or id.
func (c *Client) GetBerry(ctx context.Context, nameOrID string) (*Berry, error) {
	u := fmt.Sprintf("%s/berry/%s", c.cfg.BaseURL, nameOrID)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw berryResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode berry: %w", err)
	}
	return &Berry{
		ID:         raw.ID,
		Name:       raw.Name,
		GrowthTime: raw.GrowthTime,
		MaxHarvest: raw.MaxHarvest,
		Firmness:   raw.Firmness.Name,
		GiftType:   raw.NaturalGiftType.Name,
		GiftPower:  raw.NaturalGiftPower,
	}, nil
}

// GetItem fetches the item record by name or id.
func (c *Client) GetItem(ctx context.Context, nameOrID string) (*Item, error) {
	u := fmt.Sprintf("%s/item/%s", c.cfg.BaseURL, nameOrID)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw itemResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode item: %w", err)
	}
	effect, _ := firstEnglishEffect(raw.EffectEntries)
	return &Item{
		ID:         raw.ID,
		Name:       raw.Name,
		Cost:       raw.Cost,
		FlingPower: raw.FlingPower,
		Category:   raw.Category.Name,
		Effect:     effect,
	}, nil
}

// --- HTTP helpers ---

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return b, err != nil, err
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		return 5 * time.Second
	}
	return d
}

// --- helpers ---

func firstEnglishFlavor(entries []flavorEntry) string {
	r := strings.NewReplacer("\n", " ", "\f", " ")
	for _, e := range entries {
		if e.Language.Name == "en" {
			return r.Replace(e.FlavorText)
		}
	}
	return ""
}

func firstEnglishEffect(entries []effectEntry) (effect, short string) {
	for _, e := range entries {
		if e.Language.Name == "en" {
			return e.Effect, e.ShortEffect
		}
	}
	return "", ""
}

func nameURLSlice(refs []nameURL) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, r.Name)
	}
	return out
}
