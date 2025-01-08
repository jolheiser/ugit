{
  pkgs ? import <nixpkgs>,
}:
let
  name = "ugitd";
in
pkgs.buildGoModule {
  pname = name;
  version = "main";
  src = pkgs.nix-gitignore.gitignoreSource [ ] (
    builtins.path {
      inherit name;
      path = ../.;
    }
  );
  subPackages = [
    "cmd/ugitd"
    "cmd/ugit-uci"
  ];
  vendorHash = pkgs.lib.fileContents ../go.mod.sri;
  env.CGO_ENABLED = 0;
  flags = [ "-trimpath" ];
  ldflags = [
    "-s"
    "-w"
    "-extldflags -static"
  ];
  meta = {
    description = "Minimal git server";
    homepage = "https://git.jolheiser.com/ugit";
    mainProgram = "ugitd";
  };
}
