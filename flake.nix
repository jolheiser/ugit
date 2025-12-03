{
  description = "Minimal git server";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    tailwind-ctp = {
      url = "git+https://git.jolheiser.com/tailwind-ctp";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    tailwind-ctp-lsp = {
      url = "git+https://git.jolheiser.com/tailwind-ctp-intellisense";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      tailwind-ctp,
      tailwind-ctp-lsp,
    }:
    let
      systems = [
        "x86_64-linux"
        "i686-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "armv6l-linux"
        "armv7l-linux"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems f;
      tctp = forAllSystems (system: tailwind-ctp.packages.${system}.default);
      tctpl = forAllSystems (system: tailwind-ctp-lsp.packages.${system}.default);
    in
    {
      packages = forAllSystems (system: import ./nix { pkgs = nixpkgs.legacyPackages.${system}; });
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            nativeBuildInputs = with pkgs; [
              go
              gopls
              air
              tctp.${system}
              #tctpl.${system}
              #vscode-langservers-extracted
            ];
          };
        }
      );
      nixosModules.default = import ./nix/module.nix;
      nixosConfigurations.ugitVM = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [
          ./nix/vm.nix
          {
            virtualisation.vmVariant.virtualisation = {
              cores = 2;
              memorySize = 2048;
              graphics = false;
            };
            system.stateVersion = "23.11";
          }
        ];
      };
      apps = forAllSystems (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          vm = {
            type = "app";
            program = "${pkgs.writeShellScript "vm" ''
              nixos-rebuild build-vm --flake .#ugitVM
              ./result/bin/run-nixos-vm
              rm nixos.qcow2
            ''}";
          };
        }
      );
    };
}
