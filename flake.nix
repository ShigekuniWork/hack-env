{
  description = "pentest environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        cupp = pkgs.stdenv.mkDerivation {
            pname = "cupp";
            version = "3.3.1";

            src = pkgs.fetchFromGitHub {
              owner = "Mebus";
              repo = "cupp";
              rev = "master";
              hash = "sha256-eNE8DUFtFFEfBsm3TrL3GcnBXPQ7x0kfzZHH49Jqy5w=";
            };

            nativeBuildInputs = [
              pkgs.makeWrapper
            ];

            installPhase = ''
              mkdir -p $out/bin $out/share/cupp

              cp cupp.py $out/share/cupp/
              cp cupp.cfg $out/share/cupp/

              makeWrapper ${pkgs.python3}/bin/python $out/bin/cupp \
                --add-flags "$out/share/cupp/cupp.py"
            '';
          };
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            nmap
            ffuf
            curl
            jq
            ripgrep
            python3
            go
            git
            thc-hydra
            cupp
            ruby
          ];
        };
      });
}