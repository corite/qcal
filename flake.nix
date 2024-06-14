{
  # based on https://github.com/NixOS/templates/blob/master/go-hello/flake.nix
  description = "CLI calendar application for CalDAV servers written in Go";
  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = {
    self,
    nixpkgs,
  }: let
    lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
    version = "0.9.2";
    supportedSystems = ["x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin"];
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    nixpkgsFor = forAllSystems (system: import nixpkgs {inherit system;});
  in {
    packages = forAllSystems (system: let
      pkgs = nixpkgsFor.${system};
    in {
      qcal = pkgs.buildGoModule {
        pname = "qcal";
        inherit version;
        src = ./.;
        # vendorHash = pkgs.lib.fakeHash;
        vendorHash = "sha256-W9g2JzShvm2hJ+fcdwsoD3B6iUU55ufN6FTTl6qK6Oo=";
      };
    });
    devShells = forAllSystems (system: let
      pkgs = nixpkgsFor.${system};
    in {
      default = pkgs.mkShell {
        buildInputs = with pkgs; [go gopls gotools go-tools];
      };
    });
    defaultPackage = forAllSystems (system: self.packages.${system}.qcal);
  };
}
