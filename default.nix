{ pkgs ? import <nixpkgs> {} }:
pkgs.buildGoModule rec {
  pname = "smtp-forward";
  version = "0.0.1";

  src = ./.;
  vendorSha256 = "sha256-Aqi8Kkh3BWWyeGeyrHJxikhsmMWxpYSwixNsQyTI1R0=";

  doCheck = true;
  checkInputs = [ pkgs.go-tools pkgs.gotools ];
  checkPhase = ''
    go test ./...
    go vet ./...
    shadow ./...
    export STATICCHECK_CACHE=$PWD/staticcheck-cache
    staticcheck ./...
  '';
}
