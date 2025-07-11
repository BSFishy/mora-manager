{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ.url = "github:a-h/templ";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      templ,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells = {
          default = pkgs.mkShell {
            buildInputs = [
              pkgs.go
              pkgs.k9s
              pkgs.just
              pkgs.tilt
              templ.packages.${system}.templ
            ];
          };
        };
      }
    );
}
