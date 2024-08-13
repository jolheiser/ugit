{
  pkgs ? import <nixpkgs>,
}:
let
  pkg = pkgs.callPackage ./pkg.nix { inherit pkgs; };
in
{
  ugit = pkg;
  default = pkg;
}
