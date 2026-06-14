package pokeapi

// PokemonListItem is one row from the /pokemon list endpoint.
type PokemonListItem struct {
	Rank int    `json:"rank"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Pokemon is the full detail record for one Pokémon.
type Pokemon struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	Height         int      `json:"height"`         // decimeters
	Weight         int      `json:"weight"`         // hectograms
	BaseExperience int      `json:"base_experience"`
	Types          []string `json:"types"`
	Abilities      []string `json:"abilities"`
	HP             int      `json:"hp"`
	Attack         int      `json:"attack"`
	Defense        int      `json:"defense"`
	SpAttack       int      `json:"sp_attack"`
	SpDefense      int      `json:"sp_defense"`
	Speed          int      `json:"speed"`
	SpriteURL      string   `json:"sprite_url"` // front_default sprite
	URL            string   `json:"url"`        // https://pokeapi.co/api/v2/pokemon/{id}/
}

// Species is the Pokémon species record from /pokemon-species.
type Species struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	IsLegendary   bool   `json:"is_legendary"`
	IsMythical    bool   `json:"is_mythical"`
	CaptureRate   int    `json:"capture_rate"`
	BaseHappiness int    `json:"base_happiness"`
	Color         string `json:"color"`
	Generation    string `json:"generation"`
	FlavorText    string `json:"flavor_text"` // first English entry
}

// Type is the Pokémon type info from /type.
type Type struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	PokemonCount     int      `json:"pokemon_count"`    // len(pokemon)
	SuperEffective   []string `json:"super_effective"`  // double_damage_to names
	WeakTo           []string `json:"weak_to"`          // double_damage_from names
	NotVeryEffective []string `json:"not_very_effective"` // half_damage_to names
	NoEffect         []string `json:"no_effect"`        // no_damage_to names
}

// Ability is the ability detail from /ability.
type Ability struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PokemonCount int    `json:"pokemon_count"` // len(pokemon)
	Effect       string `json:"effect"`        // first English effect
	ShortEffect  string `json:"short_effect"`  // first English short_effect
}

// Move is the move detail from /move.
type Move struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Power       int    `json:"power"`
	Accuracy    int    `json:"accuracy"`
	PP          int    `json:"pp"`
	Type        string `json:"type"`
	DamageClass string `json:"damage_class"`
	Effect      string `json:"effect"` // first English effect
}

// Generation is the generation detail from /generation.
type Generation struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	MainRegion   string `json:"main_region"`
	PokemonCount int    `json:"pokemon_count"` // len(pokemon_species)
	TypeCount    int    `json:"type_count"`    // len(types)
}

// Item is the item detail from /item.
type Item struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Cost       int    `json:"cost"`
	FlingPower int    `json:"fling_power"`
	Category   string `json:"category"`
	Effect     string `json:"effect"` // first English effect
}

// --- unexported wire types used for JSON decode only ---

type listResponse struct {
	Count   int              `json:"count"`
	Next    string           `json:"next"`
	Results []listResultItem `json:"results"`
}

type listResultItem struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type pokemonResponse struct {
	ID             int               `json:"id"`
	Name           string            `json:"name"`
	BaseExperience int               `json:"base_experience"`
	Height         int               `json:"height"`
	Weight         int               `json:"weight"`
	Types          []pokemonTypeSlot `json:"types"`
	Abilities      []abilitySlot     `json:"abilities"`
	Stats          []statSlot        `json:"stats"`
	Sprites        sprites           `json:"sprites"`
}

type pokemonTypeSlot struct {
	Slot int     `json:"slot"`
	Type nameURL `json:"type"`
}

type abilitySlot struct {
	Ability  nameURL `json:"ability"`
	IsHidden bool    `json:"is_hidden"`
}

type statSlot struct {
	BaseStat int     `json:"base_stat"`
	Stat     nameURL `json:"stat"`
}

type sprites struct {
	FrontDefault string `json:"front_default"`
}

type speciesResponse struct {
	ID                int           `json:"id"`
	Name              string        `json:"name"`
	IsLegendary       bool          `json:"is_legendary"`
	IsMythical        bool          `json:"is_mythical"`
	CaptureRate       int           `json:"capture_rate"`
	BaseHappiness     int           `json:"base_happiness"`
	Color             nameURL       `json:"color"`
	Generation        nameURL       `json:"generation"`
	FlavorTextEntries []flavorEntry `json:"flavor_text_entries"`
}

type flavorEntry struct {
	FlavorText string  `json:"flavor_text"`
	Language   nameURL `json:"language"`
}

type typeResponse struct {
	ID              int           `json:"id"`
	Name            string        `json:"name"`
	Pokemon         []typePokemon `json:"pokemon"`
	DamageRelations typeRelations `json:"damage_relations"`
}

type typePokemon struct {
	Pokemon nameURL `json:"pokemon"`
	Slot    int     `json:"slot"`
}

type typeRelations struct {
	DoubleDamageTo   []nameURL `json:"double_damage_to"`
	DoubleDamageFrom []nameURL `json:"double_damage_from"`
	HalfDamageTo     []nameURL `json:"half_damage_to"`
	NoDamageTo       []nameURL `json:"no_damage_to"`
}

type abilityResponse struct {
	ID            int              `json:"id"`
	Name          string           `json:"name"`
	Pokemon       []abilityPokemon `json:"pokemon"`
	EffectEntries []effectEntry    `json:"effect_entries"`
}

type abilityPokemon struct {
	Pokemon  nameURL `json:"pokemon"`
	IsHidden bool    `json:"is_hidden"`
}

type effectEntry struct {
	Effect      string  `json:"effect"`
	ShortEffect string  `json:"short_effect"`
	Language    nameURL `json:"language"`
}

type moveResponse struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Power         int           `json:"power"`
	Accuracy      int           `json:"accuracy"`
	PP            int           `json:"pp"`
	Type          nameURL       `json:"type"`
	DamageClass   nameURL       `json:"damage_class"`
	EffectEntries []effectEntry `json:"effect_entries"`
}

type generationResponse struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	MainRegion     nameURL   `json:"main_region"`
	PokemonSpecies []nameURL `json:"pokemon_species"`
	Types          []nameURL `json:"types"`
}

type itemResponse struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Cost          int           `json:"cost"`
	FlingPower    int           `json:"fling_power"`
	Category      nameURL       `json:"category"`
	EffectEntries []effectEntry `json:"effect_entries"`
}

type nameURL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
