package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadRepositoryDependencyGraph(t *testing.T) {
	// TODO: use an inline text
	graph, err := LoadRepositoryGraphFromUrl("https://moab-filer.crto.in/csharp/moabs/78246/dependency-graph.json")
	assert.Nil(t, err)
	assert.NotNil(t, graph)
	assert.True(t, len(graph.Repositories) > 0)

	dependencies := graph.Dependencies([]string{"platform/mapi"}, false)
	fmt.Println(dependencies)
	assert.True(t, len(dependencies) > 0)
}
