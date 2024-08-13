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
      packages = forAllSystems (system: import ./nix { pkgs = import nixpkgs { inherit system; }; });
      devShells = forAllSystems (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = pkgs.mkShell {
            nativeBuildInputs = with pkgs; [
              go
              gopls
              templ
              tctp.${system}
              tctpl.${system}
              vscode-langservers-extracted
            ];
          };
        }
      );
      nixosModules.default = import ./nix/module.nix;
    };
}
