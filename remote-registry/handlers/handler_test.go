package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/criteo/command-launcher/internal/remote"
	"github.com/criteo/command-launcher/remote-registry/model"
	"github.com/criteo/command-launcher/remote-registry/store"
	"github.com/stretchr/testify/assert"
)

func TestHomeHandlers(t *testing.T) {
	// Create a new in-memory store
	s := store.NewInMemoryStore()

	// Create a controller instance by specifying the store
	controller := NewController(s)

	// Create a new HTTP request to test the HomePageHandler
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %s", err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the HomePageHandler with the request and recorder
	controller.HomePageHandler(rr, req)

	status, regs, err := getEntireRegistryContent(controller)

	// Check the response status code
	if status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	assert.NoError(t, err, "Failed to parse response body")
	assert.Equal(t, 0, len(regs), "Expected 1 registry")
}

func TestRegistryHandlers(t *testing.T) {
	// Create a new in-memory store
	s := store.NewInMemoryStore()

	// Create a controller instance by specifying the store
	controller := NewController(s)

	// Create a new HTTP request to test the NewRegistryHandler
	// Here we are using a POST request to create a new registry
	// create a new registry body
	registryBody := `{
		"name": "test-registry", 
		"description": "Test Registry", 
		"admin": ["a", "b"], 
		"customValues": {"key": "value"}
	}`

	req, err := http.NewRequest("POST", "/registry", bytes.NewReader([]byte(registryBody)))
	if err != nil {
		t.Fatalf("Failed to create request: %s", err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the NewRegistryHandler with the request and recorder
	controller.NewRegistryHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// check entire registry content
	_, regs, _ := getEntireRegistryContent(controller)

	assert.Equal(t, 1, len(regs), "Expected 1 registry")
	assert.Equal(t, "test-registry", regs[0].Name, "Expected registry name to be 'test-registry'")
	assert.Equal(t, "Test Registry", regs[0].Description, "Expected registry description to be 'Test Registry'")
	assert.Equal(t, []string{"a", "b"}, regs[0].Admin, "Expected registry admin to be ['a', 'b']")
	assert.Equal(t, map[string]string{"key": "value"}, regs[0].CustomValues, "Expected registry custom values to be {'key': 'value'}")

	// now delete the registry
	req, _ = http.NewRequest("DELETE", "/registry/{registry}", nil)
	req.SetPathValue("registry", "test-registry")
	rr = httptest.NewRecorder()

	controller.UpdateOrDeleteRegistryHandler(rr, req)
	// Check the response status code
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// check entire registry content
	_, regs, _ = getEntireRegistryContent(controller)
	assert.Equal(t, 0, len(regs), "Expected 0 registry")
}

func TestPackageHandlers(t *testing.T) {
	s := store.NewInMemoryStore()
	initStore(s)
	controller := NewController(s)

	// first check if the registry is correctly initiated
	_, regs, _ := getEntireRegistryContent(controller)
	assert.Equal(t, 1, len(regs), "Expected 1 registry")
	assert.Equal(t, "test-registry", regs[0].Name, "Expected registry name to be 'test-registry'")
	assert.Equal(t, "Test Registry", regs[0].Description, "Expected registry description to be 'Test Registry'")
	assert.Equal(t, []string{"a", "b"}, regs[0].Admin, "Expected registry admin to be ['a', 'b']")
	assert.Equal(t, map[string]string{"key": "value"}, regs[0].CustomValues, "Expected registry custom values to be {'key': 'value'}")
	assert.Equal(t, 1, len(regs[0].Packages), "Expected 1 package")
	assert.Equal(t, "test-package", regs[0].Packages["test-package"].Name, "Expected package name to be 'test-package'")
	assert.Equal(t, "Test Package", regs[0].Packages["test-package"].Description, "Expected package description to be 'Test Package'")
	assert.Equal(t, []string{"a", "b"}, regs[0].Packages["test-package"].Admin, "Expected package admin to be ['a', 'b']")
	assert.Equal(t, map[string]string{"key": "value"}, regs[0].Packages["test-package"].CustomValues, "Expected package custom values to be {'key': 'value'}")
	assert.Equal(t, 1, len(regs[0].Packages["test-package"].Versions), "Expected 1 package version")
	assert.Equal(t, "1.0.0", regs[0].Packages["test-package"].Versions[0].Version, "Expected package version to be '1.0.0'")

	// Create a new HTTP request to test the NewPackageHandler
	packageBody := `{
		"name": "test-package-2",
		"description": "Test Package 2",
		"admin": ["a", "b"],
		"customValues": {"key": "value"}
	}`
	req, err := http.NewRequest("POST", "/registry/{registry}/package", bytes.NewReader([]byte(packageBody)))
	if err != nil {
		t.Fatalf("Failed to create request: %s", err)
	}
	req.SetPathValue("registry", "test-registry")
	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()
	// Call the NewPackageHandler with the request and recorder
	controller.NewPackageHandler(rr, req)
	// Check the response status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// check entire registry content
	_, regs, _ = getEntireRegistryContent(controller)
	assert.Equal(t, 1, len(regs), "Expected 1 registry")
	assert.Equal(t, 2, len(regs[0].Packages), "Expected 2 packages")
	assert.Equal(t, "test-package", regs[0].Packages["test-package"].Name, "Expected package name to be 'test-package'")
	assert.Equal(t, "test-package-2", regs[0].Packages["test-package-2"].Name, "Expected package name to be 'test-package-2'")
	assert.Equal(t, "Test Package 2", regs[0].Packages["test-package-2"].Description, "Expected package description to be 'Test Package 2'")
	assert.Equal(t, []string{"a", "b"}, regs[0].Packages["test-package-2"].Admin, "Expected package admin to be ['a', 'b']")
	assert.Equal(t, map[string]string{"key": "value"}, regs[0].Packages["test-package-2"].CustomValues, "Expected package custom values to be {'key': 'value'}")

	// now delete the package
	req, _ = http.NewRequest("DELETE", "/registry/{registry}/package/{package}", nil)
	req.SetPathValue("registry", "test-registry")
	req.SetPathValue("package", "test-package-2")
	rr = httptest.NewRecorder()
	controller.UpdateOrDeletePackageHandler(rr, req)
	// Check the response status code
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}

	// check entire registry content again
	_, regs, _ = getEntireRegistryContent(controller)
	assert.Equal(t, 1, len(regs), "Expected 1 registry")
	assert.Equal(t, "test-registry", regs[0].Name, "Expected registry name to be 'test-registry'")
	assert.Equal(t, "Test Registry", regs[0].Description, "Expected registry description to be 'Test Registry'")
	assert.Equal(t, []string{"a", "b"}, regs[0].Admin, "Expected registry admin to be ['a', 'b']")
	assert.Equal(t, map[string]string{"key": "value"}, regs[0].CustomValues, "Expected registry custom values to be {'key': 'value'}")
	assert.Equal(t, 1, len(regs[0].Packages), "Expected 1 package")
	assert.Equal(t, "test-package", regs[0].Packages["test-package"].Name, "Expected package name to be 'test-package'")
	assert.Equal(t, "Test Package", regs[0].Packages["test-package"].Description, "Expected package description to be 'Test Package'")
	assert.Equal(t, []string{"a", "b"}, regs[0].Packages["test-package"].Admin, "Expected package admin to be ['a', 'b']")
	assert.Equal(t, map[string]string{"key": "value"}, regs[0].Packages["test-package"].CustomValues, "Expected package custom values to be {'key': 'value'}")
	assert.Equal(t, 1, len(regs[0].Packages["test-package"].Versions), "Expected 1 package version")
	assert.Equal(t, "1.0.0", regs[0].Packages["test-package"].Versions[0].Version, "Expected package version to be '1.0.0'")
}

func TestPackageVersionHandlers(t *testing.T) {
	s := store.NewInMemoryStore()
	initStore(s)
	controller := NewController(s)

	// first check if the registry is correctly initiated
	_, regs, _ := getEntireRegistryContent(controller)
	assert.Equal(t, 1, len(regs), "Expected 1 registry")
	assert.Equal(t, "test-registry", regs[0].Name, "Expected registry name to be 'test-registry'")
	assert.Equal(t, "Test Registry", regs[0].Description, "Expected registry description to be 'Test Registry'")
	assert.Equal(t, []string{"a", "b"}, regs[0].Admin, "Expected registry admin to be ['a', 'b']")
	assert.Equal(t, map[string]string{"key": "value"}, regs[0].CustomValues, "Expected registry custom values to be {'key': 'value'}")
	assert.Equal(t, 1, len(regs[0].Packages), "Expected 1 package")
	assert.Equal(t, "test-package", regs[0].Packages["test-package"].Name, "Expected package name to be 'test-package'")
	assert.Equal(t, "Test Package", regs[0].Packages["test-package"].Description, "Expected package description to be 'Test Package'")
	assert.Equal(t, []string{"a", "b"}, regs[0].Packages["test-package"].Admin, "Expected package admin to be ['a', 'b']")
	assert.Equal(t, map[string]string{"key": "value"}, regs[0].Packages["test-package"].CustomValues, "Expected package custom values to be {'key': 'value'}")
	assert.Equal(t, 1, len(regs[0].Packages["test-package"].Versions), "Expected 1 package version")
	assert.Equal(t, "1.0.0", regs[0].Packages["test-package"].Versions[0].Version, "Expected package version to be '1.0.0'")

	// Create a new HTTP request to test the NewPackageVersionHandler
	packageVersionBody := `{
		"name": "test-package",
		"version": "1.1.0",
		"url": "http://example.com/test-package-1.1.0.tar.gz",
		"checksum": "abc456",
		"startPartition": 0,
		"endPartition": 9
	}`

	req, err := http.NewRequest("POST", "/registry/{registry}/package/{package}/version", bytes.NewReader([]byte(packageVersionBody)))
	if err != nil {
		t.Fatalf("Failed to create request: %s", err)
	}
	req.SetPathValue("registry", "test-registry")
	req.SetPathValue("package", "test-package")
	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the NewPackageVersionHandler with the request and recorder
	controller.NewPackageVersionHandler(rr, req)
	// Check the response status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
	// check newly created version
	_, regs, _ = getEntireRegistryContent(controller)
	assert.Equal(t, 1, len(regs), "Expected 1 registry")
	assert.Equal(t, 1, len(regs[0].Packages), "Expected 1 package")
	assert.Equal(t, 2, len(regs[0].Packages["test-package"].Versions), "Expected 2 package versions")
	assert.Equal(t, "1.0.0", regs[0].Packages["test-package"].Versions[0].Version, "Expected package version to be '1.0.0'")
	assert.Equal(t, "1.1.0", regs[0].Packages["test-package"].Versions[1].Version, "Expected package version to be '1.1.0'")

	// now delete the package version
	req, _ = http.NewRequest("DELETE", "/registry/{registry}/package/{package}/version/{version}", nil)
	req.SetPathValue("registry", "test-registry")
	req.SetPathValue("package", "test-package")
	req.SetPathValue("version", "1.0.0")
	rr = httptest.NewRecorder()
	controller.DeletePackageVersionHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}

	// check entire registry content again
	_, regs, _ = getEntireRegistryContent(controller)
	assert.Equal(t, 1, len(regs), "Expected 1 registry")
	assert.Equal(t, 1, len(regs[0].Packages), "Expected 1 package")
	assert.Equal(t, 1, len(regs[0].Packages["test-package"].Versions), "Expected 1 package version")
	assert.Equal(t, "1.1.0", regs[0].Packages["test-package"].Versions[0].Version, "Expected package version to be '1.1.0'")
}

/**
 * Helpers comes here
 */
func initStore(s store.Store) {
	// put some example data in the store
	s.NewRegistry("test-registry", model.RegistryMetadata{
		Name:         "test-registry",
		Description:  "Test Registry",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	})

	s.NewPackage("test-registry", "test-package", model.PackageMetadata{
		Name:         "test-package",
		Description:  "Test Package",
		Admin:        []string{"a", "b"},
		CustomValues: map[string]string{"key": "value"},
	})
	s.NewPackageVersion("test-registry", "test-package", "1.0.0", remote.PackageInfo{
		Name:           "test-package",
		Version:        "1.0.0",
		Url:            "http://example.com/test-package-1.0.0.tar.gz",
		Checksum:       "abc123",
		StartPartition: 0,
		EndPartition:   9,
	})
}

func getEntireRegistryContent(c *Controller) (int, []model.Registry, error) {
	// Create a new HTTP request to test the HomePageHandler
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return 500, []model.Registry{}, err
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the HomePageHandler with the request and recorder
	c.HomePageHandler(rr, req)

	regs, err := parseEntireRegistries(rr.Body.String())

	return rr.Result().StatusCode, regs, err
}

func parseEntireRegistries(s string) ([]model.Registry, error) {
	var regs []model.Registry
	err := json.Unmarshal([]byte(s), &regs)
	if err != nil {
		return nil, err
	}
	return regs, nil
}
