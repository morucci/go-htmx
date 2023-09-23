{
  description = "go-htmx playground";
  nixConfig.bash-prompt = "[nix(go-htmx)] ";
  inputs = { nixpkgs.url = "github:nixos/nixpkgs/23.05"; };

  outputs = { self, nixpkgs }:
    let pkgs = nixpkgs.legacyPackages.x86_64-linux.pkgs;
    in {
      devShells.x86_64-linux.default = pkgs.mkShell {
        name = "go-htmx dev shell";
        buildInputs = [
          # 1.20.4 in nixpkgs
          pkgs.go
          # 0.11.0 in nixpkgs
          pkgs.gopls
        ];
        shellHook = ''
          echo "Welcome in $name"
          export GOPATH=$(go env GOPATH)
          export GOBIN=$(go env GOPATH)/bin
        '';
      };
    };
}
