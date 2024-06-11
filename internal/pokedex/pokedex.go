package pokedex

import "sync"

type Pokemon struct {
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Name           string `json:"name"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
}

type LocationArea struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type LocationAreas struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type Pokedex struct {
	pokemon map[string]Pokemon
	mu      sync.Mutex
}

func New() *Pokedex {
	return &Pokedex{
		pokemon: make(map[string]Pokemon),
	}
}

func (p *Pokedex) AddPokemon(pokemonName string, pokemon Pokemon) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pokemon[pokemonName] = pokemon
}

func (p *Pokedex) GetPokemon(pokemonName string) (Pokemon, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pokemon, ok := p.pokemon[pokemonName]
	if !ok {
		return Pokemon{}, false
	}

	return pokemon, true
}

func (p *Pokedex) GetAllPokemonNames() []string {
	pokemonNames := []string{}
	for pokemonName := range p.pokemon {
		pokemonNames = append(pokemonNames, pokemonName)
	}

	return pokemonNames
}
