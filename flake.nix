{
  description = "Nix flake to create a dev environment for the Akri Discovery Handler in Go";
  inputs = 
  {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs = { self, nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in 
    {
      devShells.x86_64-linux.default = pkgs.mkShell { 
        nativeBuildInputs = with pkgs; [
          go
          git
        ];
      };
    };
}