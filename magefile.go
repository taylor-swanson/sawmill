//go:build mage
// +build mage

package main

import "github.com/magefile/mage/sh"

// Builds the project
func Build() error {
	var err error

	if err = sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	if err = sh.Run("go", "build", "-o", "build/sawmill", "github.com/taylor-swanson/sawmill/cmd/sawmill"); err != nil {
		return err
	}

	return nil
}

// Generates project files
func Generate() error {
	var err error

	genPaths := []string{}

	for _, v := range genPaths {
		if err = sh.Run("go", "generate", v); err != nil {
			return err
		}
	}

	return nil
}

// Cleans build artifacts
func Clean() error {
	var err error

	if err = sh.Run("rm", "-rf", "build/"); err != nil {
		return err
	}

	return nil
}

// Build everything
func All() error {
	var err error

	if err = Generate(); err != nil {
		return err
	}
	if err = Build(); err != nil {
		return err
	}

	return nil
}
