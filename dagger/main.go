package main

import "context"

type Ugit struct{}

// Base nix container
func (u *Ugit) Nix(source *Directory) *Container {
	return dag.Container().
		From("nixos/nix:latest").
		WithDirectory("/src", source).
		WithWorkdir("/src")
}

// Nix build
func (u *Ugit) Build(source *Directory) *Container {
	return u.Nix(source).
		WithExec([]string{
			"nix",
			"--experimental-features",
			"nix-command flakes",
			"build",
		})
}

// Push to cachix
func (u *Ugit) Cachix(ctx context.Context, source *Directory, cachix *Secret) (string, error) {
	return u.Build(source).
		WithSecretVariable("CACHIX_AUTH_TOKEN", cachix).
		WithExec([]string{
			"nix",
			"--experimental-features",
			"nix-command flakes",
			"run",
			"nixpkgs#cachix",
			"--",
			"push",
			"jolheiser",
			"./result",
		}).
		Stdout(ctx)
}
