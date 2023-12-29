{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";
    nur = {
      url = "git+https://git.jojodev.com/jolheiser/nur";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    tailwind-ctp = {
      url = "git+https://git.jojodev.com/jolheiser/tailwind-ctp";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    flake-utils,
    nixpkgs,
    nur,
    tailwind-ctp,
  } @ inputs: let
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};
    nur = inputs.nur.packages.${system};
    tailwind-ctp = inputs.tailwind-ctp.packages.${system}.default;
  in {
    devShells.${system}.default = pkgs.mkShell {
      nativeBuildInputs = with pkgs; [
        # go
        # gopls
        nur.templ
        tailwind-ctp
        sqlc
      ];
    };
  };
}
