package model

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/criteo/command-launcher/internal/helper"
)

type RepositoryGraph struct {
	Repositories map[string]Repository
}
type Repository struct {
	Name                 string   `json:"name"`
	RequiredRepositories []string `json:"requiredProjects"`
	ClientRepositories   []string
}

func LoadRepositoryGraphFromFile(file string) (RepositoryGraph, error) {
	graph := RepositoryGraph{}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return graph, err
	}
	return LoadRepositoryGraph(data)
}

func LoadRepositoryGraphFromUrl(url string) (RepositoryGraph, error) {
	graph := RepositoryGraph{}
	resp, err := helper.HttpGetWrapper(url)
	if err != nil {
		return graph, err
	}

	if !helper.Is2xx(resp.StatusCode) {
		return graph, fmt.Errorf("failed to repository dependency graph from url: %s", url)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return graph, err
	}
	return LoadRepositoryGraph(body)
}

func LoadRepositoryGraph(graphText []byte) (RepositoryGraph, error) {
	graph := RepositoryGraph{
		Repositories: map[string]Repository{},
	}
	repositories := []Repository{}
	err := json.Unmarshal([]byte(graphText), &repositories)

	for _, repo := range repositories {
		graph.Repositories[repo.Name] = repo
	}

	return graph, err
}

func (g RepositoryGraph) Dependencies(repositories []string, transit bool) []string {
	depMap := map[string]bool{}

	for _, repository := range repositories {
		g.getDependencies(repository, transit, depMap, map[string]bool{})
	}

	output := make([]string, 0, len(depMap))
	for k := range depMap {
		output = append(output, k)
	}
	return output
}

func (g RepositoryGraph) getDependencies(repository string, transit bool, output map[string]bool, visited map[string]bool) {
	if _, ok := g.Repositories[repository]; !ok {
		return
	}
	repo := g.Repositories[repository]

	if _, ok := visited[repository]; ok {
		return
	}
	visited[repository] = true

	for _, requirement := range repo.RequiredRepositories {
		if _, ok := visited[requirement]; !ok {
			output[requirement] = true
		}
	}

	if transit {
		for _, direct := range g.Repositories[repository].RequiredRepositories {
			g.getDependencies(direct, transit, output, visited)
		}
	}
}
