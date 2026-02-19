{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      nixpkgs,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = nixpkgs.lib.systems.flakeExposed;
      perSystem =
        { system, ... }:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [
              (final: prev: {
                go = prev.go_1_26;
                buildGoModule = prev.buildGoModule.override { go = prev.go_1_26; };
              })
            ];
          };
          inherit (pkgs) lib callPackage;
          version = self.shortRev or self.dirtyShortRev or "snapshot";
          scripts = lib.packagesFromDirectoryRecursive {
            inherit callPackage;
            directory = ./nix/scripts;
          };
          llm-cli = callPackage ./nix/package.nix {
            inherit version;
          };
          shell = callPackage ./nix/shell.nix { };
        in
        {
          packages = scripts // {
            default = llm-cli;
          };
          devShells = {
            default = shell;
          };
        };
    };
}
