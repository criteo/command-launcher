package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	remote "github.com/criteo/command-launcher/internal/remote"
	model "github.com/criteo/command-launcher/remote-registry/model"
)

type S3StoreConfig struct {
	Bucket          string
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
}

type S3Store struct {
	client          *s3.Client
	bucket          string
	mutex           sync.RWMutex
	registriesCache map[string]*model.Registry
}

func NewS3Store(cfg S3StoreConfig) (*S3Store, error) {
	ctx := context.Background()

	// MinIO/S3-compatible services don't use regions, but AWS SDK v2 requires a region parameter.
	// If a custom endpoint is provided (indicating MinIO) and no region is specified,
	// we automatically use a placeholder value that MinIO will ignore.
	region := cfg.Region
	if region == "" && cfg.Endpoint != "" {
		region = "us-east-1" // Placeholder value - MinIO ignores this but SDK requires it
	}
	if region == "" {
		return nil, fmt.Errorf("s3-region is required for AWS S3 (or provide s3-endpoint for S3-compatible services)")
	}

	var awsConfig aws.Config
	var err error

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
	} else {
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		if cfg.UsePathStyle {
			o.UsePathStyle = true
		}
	})

	store := &S3Store{
		client:          s3Client,
		bucket:          cfg.Bucket,
		registriesCache: make(map[string]*model.Registry),
	}

	if err := store.loadRegistriesFromS3(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *S3Store) getRegistryKey(name string) string {
	return fmt.Sprintf("registries/%s.json", name)
}

func (s *S3Store) loadRegistriesFromS3() error {
	ctx := context.Background()

	listOutput, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String("registries/"),
	})
	if err != nil {
		return fmt.Errorf("failed to list S3 objects: %w", err)
	}

	for _, obj := range listOutput.Contents {
		if !strings.HasSuffix(*obj.Key, ".json") {
			continue
		}

		parts := strings.Split(*obj.Key, "/")
		if len(parts) != 2 {
			continue
		}

		registryName := strings.TrimSuffix(parts[1], ".json")
		registry, err := s.loadRegistryFromS3(registryName)
		if err != nil {
			return err
		}

		s.registriesCache[registryName] = registry
	}

	return nil
}

func (s *S3Store) loadRegistryFromS3(name string) (*model.Registry, error) {
	ctx := context.Background()

	key := s.getRegistryKey(name)
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, RegistryDoesNotExistError
		}
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	var registry model.Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry JSON: %w", err)
	}

	return &registry, nil
}

func (s *S3Store) saveRegistryToS3(registry *model.Registry) error {
	ctx := context.Background()

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	key := s.getRegistryKey(registry.Name)
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("failed to put object to S3: %w", err)
	}

	return nil
}

func (s *S3Store) NewRegistry(name string, registryInfo model.RegistryMetadata) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.registriesCache[name]; exists {
		return RegistryAlreadyExistsError
	}
	if name != registryInfo.Name {
		return RegistryNameMismatchError
	}

	registry := &model.Registry{
		RegistryMetadata: registryInfo,
		Packages:         make(map[string]model.Package),
	}

	if err := s.saveRegistryToS3(registry); err != nil {
		return err
	}

	s.registriesCache[name] = registry
	return nil
}

func (s *S3Store) UpdateRegistry(name string, registryInfo model.RegistryMetadata) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[name]
	if !exists {
		return RegistryDoesNotExistError
	}
	if name != registryInfo.Name {
		return RegistryNameMismatchError
	}

	registry.Description = registryInfo.Description
	registry.Admin = registryInfo.Admin
	registry.CustomValues = registryInfo.CustomValues

	if err := s.saveRegistryToS3(registry); err != nil {
		return err
	}

	return nil
}

func (s *S3Store) DeleteRegistry(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.registriesCache[name]; !exists {
		return RegistryDoesNotExistError
	}

	ctx := context.Background()
	key := s.getRegistryKey(name)
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	delete(s.registriesCache, name)
	return nil
}

func (s *S3Store) AllRegistries() ([]model.Registry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	registries := []model.Registry{}
	for _, registry := range s.registriesCache {
		registries = append(registries, *registry)
	}

	sort.Slice(registries, func(i, j int) bool {
		return registries[i].Name < registries[j].Name
	})

	return registries, nil
}

func (s *S3Store) NewPackage(registryName string, packageName string, packageInfo model.PackageMetadata) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return RegistryDoesNotExistError
	}
	if _, exists := registry.Packages[packageName]; exists {
		return PackageAlreadyExistsError
	}
	if packageName != packageInfo.Name {
		return PackageNameMismatchError
	}

	registry.Packages[packageName] = model.Package{
		PackageMetadata: packageInfo,
		Versions:        []remote.PackageInfo{},
	}

	if err := s.saveRegistryToS3(registry); err != nil {
		return err
	}

	return nil
}

func (s *S3Store) UpdatePackage(registryName string, packageName string, packageInfo model.PackageMetadata) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return RegistryDoesNotExistError
	}
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return PackageDoesNotExistError
	}
	if packageName != packageInfo.Name {
		return PackageNameMismatchError
	}

	pkg.Description = packageInfo.Description
	pkg.Admin = packageInfo.Admin
	pkg.CustomValues = packageInfo.CustomValues
	registry.Packages[packageName] = pkg

	if err := s.saveRegistryToS3(registry); err != nil {
		return err
	}

	return nil
}

func (s *S3Store) DeletePackage(registryName string, packageName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return RegistryDoesNotExistError
	}
	if _, exists := registry.Packages[packageName]; !exists {
		return PackageDoesNotExistError
	}

	delete(registry.Packages, packageName)

	if err := s.saveRegistryToS3(registry); err != nil {
		return err
	}

	return nil
}

func (s *S3Store) AllPackagesFromRegistry(registryName string) ([]model.Package, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return nil, RegistryDoesNotExistError
	}

	packages := []model.Package{}
	for _, pkg := range registry.Packages {
		packages = append(packages, pkg)
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	return packages, nil
}

func (s *S3Store) NewPackageVersion(registryName string, packageName string, version string, packageInfo remote.PackageInfo) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return RegistryDoesNotExistError
	}
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return PackageDoesNotExistError
	}
	if packageName != packageInfo.Name {
		return PackageNameMismatchError
	}

	for _, v := range pkg.Versions {
		if v.Version == version {
			return PackageVersionAlreadyExistsError
		}
	}

	pkg.Versions = append(pkg.Versions, packageInfo)
	registry.Packages[packageName] = pkg

	if err := s.saveRegistryToS3(registry); err != nil {
		return err
	}

	return nil
}

func (s *S3Store) DeletePackageVersion(registryName string, packageName string, version string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return RegistryDoesNotExistError
	}
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return PackageDoesNotExistError
	}

	newVersions := []remote.PackageInfo{}
	exists = false
	for _, v := range pkg.Versions {
		if v.Version == version {
			exists = true
		}
		if v.Version != version {
			newVersions = append(newVersions, v)
		}
	}
	if !exists {
		return PackageVersionDoesNotExistError
	}
	pkg.Versions = newVersions
	registry.Packages[packageName] = pkg

	if err := s.saveRegistryToS3(registry); err != nil {
		return err
	}

	return nil
}

func (s *S3Store) AllPackageVersionsFromRegistry(registryName string, packageName string) ([]remote.PackageInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	registry, exists := s.registriesCache[registryName]
	if !exists {
		return nil, RegistryDoesNotExistError
	}
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return nil, PackageDoesNotExistError
	}

	versions := make([]remote.PackageInfo, len(pkg.Versions))
	copy(versions, pkg.Versions)

	return versions, nil
}
