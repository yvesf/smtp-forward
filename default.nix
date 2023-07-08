{ pkgs ? import <nixpkgs> {} }:
pkgs.buildGoModule rec {
  pname = "smtp-forward";
  version = "0.0.1";

  src = ./.;
  vendorSha256 = "sha256-Aqi8Kkh3BWWyeGeyrHJxikhsmMWxpYSwixNsQyTI1R0=";

  doCheck = true;

  nativeCheckInputs = [ pkgs.golangci-lint ];
  checkPhase = ''
    export GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache
    golangci-lint run ./...
    go test ./...
  '';
}
