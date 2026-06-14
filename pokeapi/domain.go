package pokeapi

import (
	"context"
	"fmt"
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes pokeapi as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/pokeapi-cli/pokeapi"
//
// The same Domain also builds the standalone pokeapi binary (see cli.NewApp).
func init() { kit.Register(Domain{}) }

// Domain is the pokeapi driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "pokeapi",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "pokeapi",
			Short:  "Pokémon data from PokéAPI (pokeapi.co)",
			Long: `pokeapi fetches Pokémon data from the public PokéAPI. No API key required.

Retrieve the national Pokédex list, look up individual Pokémon by name or
national number, and explore types, abilities, moves, generations, and items.
Output is line-delimited JSON ready to pipe into your tools.`,
			Site: Host,
			Repo: "https://github.com/tamnd/pokeapi-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "list",
		Group:   "read",
		List:    true,
		Summary: "List Pokémon from the national Pokédex",
	}, listOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "get",
		Group:   "read",
		Single:  true,
		Summary: "Get Pokémon info (types, stats, abilities)",
		Args:    []kit.Arg{{Name: "name-or-id", Help: "Pokémon name (pikachu) or national number (25)"}},
	}, getOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "species",
		Group:   "read",
		Single:  true,
		Summary: "Get Pokémon species info (legendary, color, flavor text)",
		Args:    []kit.Arg{{Name: "name-or-id", Help: "Pokémon name or national number"}},
	}, speciesOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "type",
		Group:   "read",
		Single:  true,
		Summary: "Get type info and damage relations",
		Args:    []kit.Arg{{Name: "name-or-id", Help: "type name or id (e.g. fire, 10)"}},
	}, typeOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "ability",
		Group:   "read",
		Single:  true,
		Summary: "Get ability detail and effect",
		Args:    []kit.Arg{{Name: "name-or-id", Help: "ability name or id (e.g. overgrow)"}},
	}, abilityOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "move",
		Group:   "read",
		Single:  true,
		Summary: "Get move detail (power, accuracy, pp, effect)",
		Args:    []kit.Arg{{Name: "name-or-id", Help: "move name or id (e.g. flamethrower)"}},
	}, moveOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "generation",
		Group:   "read",
		Single:  true,
		Summary: "Get generation info",
		Args:    []kit.Arg{{Name: "name-or-id", Help: "generation name or id (e.g. generation-i, 1)"}},
	}, generationOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "item",
		Group:   "read",
		Single:  true,
		Summary: "Get item detail",
		Args:    []kit.Arg{{Name: "name-or-id", Help: "item name or id (e.g. potion)"}},
	}, itemOp)
}

// newClient builds the client from host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- input types ---

type listInput struct {
	Limit  int           `kit:"flag,inherit" help:"max results (default 20)"`
	Offset int           `kit:"flag" help:"starting offset (default 0)"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

type getInput struct {
	NameOrID string  `kit:"arg" help:"name or id"`
	Client   *Client `kit:"inject"`
}

// --- handlers ---

func listOp(ctx context.Context, in listInput, emit func(PokemonListItem) error) error {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	items, err := in.Client.List(ctx, limit, in.Offset)
	if err != nil {
		return mapErr(err)
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func getOp(ctx context.Context, in getInput, emit func(*Pokemon) error) error {
	if in.NameOrID == "" {
		return errs.Usage("name-or-id is required")
	}
	p, err := in.Client.GetPokemon(ctx, in.NameOrID)
	if err != nil {
		return mapErr(err)
	}
	return emit(p)
}

func speciesOp(ctx context.Context, in getInput, emit func(*Species) error) error {
	if in.NameOrID == "" {
		return errs.Usage("name-or-id is required")
	}
	s, err := in.Client.GetSpecies(ctx, in.NameOrID)
	if err != nil {
		return mapErr(err)
	}
	return emit(s)
}

func typeOp(ctx context.Context, in getInput, emit func(*Type) error) error {
	if in.NameOrID == "" {
		return errs.Usage("name-or-id is required")
	}
	t, err := in.Client.GetType(ctx, in.NameOrID)
	if err != nil {
		return mapErr(err)
	}
	return emit(t)
}

func abilityOp(ctx context.Context, in getInput, emit func(*Ability) error) error {
	if in.NameOrID == "" {
		return errs.Usage("name-or-id is required")
	}
	a, err := in.Client.GetAbility(ctx, in.NameOrID)
	if err != nil {
		return mapErr(err)
	}
	return emit(a)
}

func moveOp(ctx context.Context, in getInput, emit func(*Move) error) error {
	if in.NameOrID == "" {
		return errs.Usage("name-or-id is required")
	}
	m, err := in.Client.GetMove(ctx, in.NameOrID)
	if err != nil {
		return mapErr(err)
	}
	return emit(m)
}

func generationOp(ctx context.Context, in getInput, emit func(*Generation) error) error {
	if in.NameOrID == "" {
		return errs.Usage("name-or-id is required")
	}
	g, err := in.Client.GetGeneration(ctx, in.NameOrID)
	if err != nil {
		return mapErr(err)
	}
	return emit(g)
}

func itemOp(ctx context.Context, in getInput, emit func(*Item) error) error {
	if in.NameOrID == "" {
		return errs.Usage("name-or-id is required")
	}
	i, err := in.Client.GetItem(ctx, in.NameOrID)
	if err != nil {
		return mapErr(err)
	}
	return emit(i)
}

// --- Resolver ---

// Classify turns an input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty pokeapi reference")
	}
	return "pokemon", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "pokemon":
		return fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", id), nil
	default:
		return "", errs.Usage("pokeapi has no resource type %q", uriType)
	}
}

// mapErr converts a library error into the kit error kind.
func mapErr(err error) error {
	return err
}
