# smtp-forward
## usage

```nix
 inputs.smtp-forward = {
   url = "github:yvesf/smtp-forward";
   inputs.nixpkgs.follows = "nixpkgs";
 };
 # ..
 imports = [ smtp-forward.nixosModule ];
 services.smtp-forward = {
   enable = true;
   mapping = "prefix:target@host.tld,prefix2:target2@email";
   hostname = "domain-name.tld";
   cert = "/var/lib/acme/domain-name.tld/fullchain.pem";
   key = "/var/lib/acme/domain-name.tld/key.pem";
 };
```

