{
  inputs.nixpkgs.url = "nixpkgs/nixos-23.11";
  outputs = { self, nixpkgs }:
  with import nixpkgs { system = "x86_64-linux"; };
  {
    defaultPackage.x86_64-linux = callPackage ./. {};
    nixosModule = { config, ... }: { imports = [ ./module.nix ]; };
  };
}
