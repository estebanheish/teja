{
  description = "teja flake";

  inputs = {nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";};

  outputs = {
    self,
    nixpkgs,
  }: let
    allSystems = ["x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin"];
    forAllSystems = fn:
      nixpkgs.lib.genAttrs allSystems
      (system: fn {pkgs = import nixpkgs {inherit system;};});
  in {
    devShells = forAllSystems ({pkgs}: {
      default = pkgs.mkShell {
        name = "nix";
        packages = with pkgs; [go gopls];
      };
    });
    packages = forAllSystems (
      {pkgs}: {
        default = pkgs.buildGoModule {
          pname = "teja";
          version = "0.0.2";
          src = ./.;
          vendorHash = "sha256-AvSfxnMRrGRaTmozPwS8FEthFPJXNPoGtt41MO6J5fE=";
        };
      }
    );
  };
}
