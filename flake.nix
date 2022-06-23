{
  description =
    "A flake for setting up a build environment for terraform-provider-arplookup";

  inputs = {
    nixpkgs = { url = "github:NixOS/nixpkgs/nixos-22.05"; };
    flake-utils = { url = "github:numtide/flake-utils"; };
    devshell = { url = "github:numtide/devshell"; };
  };
  outputs = { self, nixpkgs, devshell, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ devshell.overlay ];
        };
        tpa = pkgs.buildGo118Module {
          inherit system;
          pname = "terraform-provider-arplookup";
          version = "0.3.1";

          src = ./.;

          vendorSha256 = "sha256-0LB2kkLvRra5oT+bvhYURNqTn6ZBmZWVnvukUGLpLRY=";

          doCheck = true;

          meta = with pkgs.lib;
            {
              description = "A Terraform provider that contains a datasource which looks up an IP address in a network given an interface MAC address.";
              homepage = "https://github.com/j-lgs/terraform-provider-arplookup";
              license = licenses.mpl20;
            };
        };
      in {
        defaultPackage = tpa;
        devShells = {
          default = pkgs.devshell.mkShell {
            name = "terraform-provider-arplookup";

            packages = with pkgs; [
              go_1_18
              gopls
              gcc

              terraform
            ];
            commands = [
              {
                name = "acctest";
                category = "testing";
                help = "Run acceptance tests.";
                command = "tools/pretest.sh";
              }
              {
                name = "build";
                category = "build";
                help = "Build the program for the current system";
                command = "nix build .#defaultPackage.${system}";
              }
              {
                name = "tests";
                category = "testing";
                help = "Run unit tests.";
                command = ''
                  go vet ./...
                  go run honnef.co/go/tools/cmd/staticcheck ./internal/arplookup
                  go test -v -race -vet=off ./...
                '';
              }
              {
                name = "generate";
                category = "code";
                help = "Regenerate documentation.";
                command = "go generate ./...";
              }
            ];
          };
        };
      });
}

