package unifi

// This will generate the *.generated.go files in this package for the specified
// client controller version.
//go:generate go run ../codegen/ -version-base-dir=../codegen/ latest
