package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"example.com/duncan/pokedexcli/internal/pokecache"
	"example.com/duncan/pokedexcli/internal/pokedex"
)

type command string
type callback func(...string) error

type commandDesc struct {
	command     command
	description string
}

type MapContext struct {
	Next     string
	Previous string
}

var commandDescs = []commandDesc{
	{
		command:     "pokedex",
		description: "List all pokemon you have caught",
	},
	{
		command:     "inspect",
		description: "Inspect a pokemon",
	},
	{
		command:     "catch",
		description: "Catch a pokemon",
	},
	{
		command:     "explore",
		description: "Explore a location",
	},
	{
		command:     "map",
		description: "Get the next list of locations",
	},
	{
		command:     "mapb",
		description: "Get the previous list of locations",
	},
	{
		command:     "help",
		description: "provides help",
	},
	{
		command:     "exit",
		description: "leave the pokedex",
	},
}

var callbacksByCommand = map[command]callback{
	"pokedex": commandPokedex,
	"inspect": commandInspect,
	"catch":   commandCatch,
	"explore": commandExplore,
	"map":     commandMapN,
	"mapb":    commandMapB,
	"help":    commandHelp,
	"exit":    commandExit,
}

var mapContext = MapContext{
	Next:     "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
	Previous: "",
}

var quit = false
var cache *pokecache.Cache = pokecache.NewCache(5 * time.Second)
var pdex *pokedex.Pokedex = pokedex.New()

func commandPokedex(args ...string) error {
	if len(args) != 0 {
		fmt.Printf("\nUsage: pokedex\n\n")
		return nil
	}

	fmt.Printf("Your Pokedex:\n")
	for _, pokemonName := range pdex.GetAllPokemonNames() {
		fmt.Printf("  - %s\n", pokemonName)
	}

	return nil
}

func commandInspect(args ...string) error {
	if len(args) != 1 {
		fmt.Printf("\nUsage: inspect [pokemon]\n\n")
		return nil
	}

	pokemonName := args[0]

	pokemon, ok := pdex.GetPokemon(pokemonName)
	if !ok {
		return errors.New("you have not caught that pokemon")
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Printf("Stats:\n")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  - %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Printf("Types:\n")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}

	return nil
}

func commandCatch(args ...string) error {
	if len(args) != 1 {
		fmt.Printf("\nUsage: catch [pokemon]\n\n")
		return nil
	}

	pokemonName := args[0]
	url := "https://pokeapi.co/api/v2/pokemon/" + pokemonName

	body, err := fetchData(url)
	if err != nil {
		return err
	}

	pokemon := pokedex.Pokemon{}
	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n\n", pokemonName)
	if pokemon.BaseExperience < rand.Intn(500) {
		fmt.Printf("%s was caught!\n", pokemonName)
		fmt.Printf("You may now inspect it with the inspect command.\n")
		pdex.AddPokemon(pokemonName, pokemon)
	} else {
		fmt.Printf("%s escaped!\n", pokemonName)
	}

	return nil
}

func commandExplore(args ...string) error {
	if len(args) != 1 {
		fmt.Printf("\nUsage: explore [location]\n\n")
		return nil
	}

	locationAreaName := args[0]
	url := "https://pokeapi.co/api/v2/location-area/" + locationAreaName

	body, err := fetchData(url)
	if err != nil {
		return err
	}

	locationArea := pokedex.LocationArea{}
	err = json.Unmarshal(body, &locationArea)
	if err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\n", locationAreaName)
	fmt.Printf("Found Pokemon:\n\n")
	for _, pokemonEncounter := range locationArea.PokemonEncounters {
		pokemon := pokemonEncounter.Pokemon
		fmt.Printf("  %s\n", pokemon.Name)
	}
	fmt.Printf("\n")

	return nil
}

func commandMapN(args ...string) error {
	if len(args) != 0 {
		fmt.Printf("\nUsage: map\n\n")
		return nil
	}

	if mapContext.Next == "" {
		return errors.New("no next locations")
	}

	return doMap(mapContext.Next)
}

func commandMapB(args ...string) error {
	if len(args) != 0 {
		fmt.Printf("\nUsage: mapb\n\n")
		return nil
	}

	if mapContext.Previous == "" {
		return errors.New("no previous locations")
	}

	return doMap(mapContext.Previous)
}

func doMap(url string) error {
	body, err := fetchData(url)
	if err != nil {
		return err
	}

	locationAreas := pokedex.LocationAreas{}
	err = json.Unmarshal(body, &locationAreas)
	if err != nil {
		return err
	}

	mapContext.Next = locationAreas.Next
	mapContext.Previous = locationAreas.Previous

	fmt.Printf("\n")
	for _, locationArea := range locationAreas.Results {
		fmt.Printf("  %s\n", locationArea.Name)
	}
	fmt.Printf("\n")

	return nil
}

func fetchData(url string) ([]byte, error) {
	var body []byte
	var ok bool

	if body, ok = cache.Get(url); !ok {
		res, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		if res.StatusCode > 299 { // TODO: is this a good way to do this?
			return nil, errors.New("bad request")
		}
		body, err = io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, err
		}

		cache.Add(url, body)
	} else {
		fmt.Println("Getting cached data...")
	}

	return body, nil
}

func commandHelp(args ...string) error {
	if len(args) != 0 {
		fmt.Printf("\nUsage: help\n\n")
		return nil
	}

	fmt.Printf("\nUsage\n\n")
	for _, commandDesc := range commandDescs {
		fmt.Printf("%s: %s\n", commandDesc.command, commandDesc.description)
	}
	fmt.Printf("\n")

	return nil
}

func commandExit(args ...string) error {
	if len(args) != 0 {
		fmt.Printf("\nUsage: exit\n\n")
		return nil
	}

	quit = true
	return nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for !quit {
		fmt.Print("pokedex > ")

		scanner.Scan()
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}

		command := command(fields[0])
		args := fields[1:]

		callback, ok := callbacksByCommand[command]
		if !ok {
			continue
		}

		err := callback(args...)
		if err != nil {
			fmt.Println(err)
		}
	}
}
