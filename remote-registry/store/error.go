package store

import "errors"

var (
	RegistryAlreadyExistsError       = errors.New("registry already exists")
	RegistryDoesNotExistError        = errors.New("registry does not exist")
	RegistryNameMismatchError        = errors.New("registry name mismatch")
	PackageAlreadyExistsError        = errors.New("package already exists")
	PackageDoesNotExistError         = errors.New("package does not exist")
	PackageNameMismatchError         = errors.New("package name mismatch")
	PackageVersionAlreadyExistsError = errors.New("package version already exists")
	PackageVersionDoesNotExistError  = errors.New("package version does not exist")
)
