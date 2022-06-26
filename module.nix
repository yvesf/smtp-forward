{ config, pkgs, lib, runCommand, ... }:
let
  package = pkgs.callPackage ./default.nix {};
  cfg = config.services.smtp-forward;
in {
  options.services.smtp-forward = {
    enable = lib.mkEnableOption "the smtp-forward service";
    mapping = lib.mkOption {
      type = lib.types.strMatching "([^:]+:[^:]+@[^,]+)(,[^:]+:[^:]+@[^,]+)*";
      default = "prefix1:addres@host,prefix2:address@host";
      description = "-m maps prefixes to email addresses";
    };
    listen = lib.mkOption {
      type = lib.types.str;
      default = ":25";
      description = "-l sets the address to listen on";
    };
    hostname = lib.mkOption {
      type = lib.types.str;
      default = "localhost";
      description = "-h sets the server hostname";
    };
    key = lib.mkOption {
      type = lib.types.path;
      default = null;
      description = "-k /path/to/file sets the TLS key file";
    };
    cert = lib.mkOption {
      type = lib.types.path;
      default = null;
      description = "-c /path/to/file sets the TLS certificate file";
    };
    from = lib.mkOption {
      type = lib.types.strMatching ".+@.+";
      default = "forwarder@localnet.cc";
      description = "-f local@domain.tld sets the From for the forwarded email";
    };
  };
  config = lib.mkIf config.services.smtp-forward.enable {
    systemd.services.smtp-forward = {
      description = "Run smtp-forward";
      path = [ package ];
      wantedBy = [ "default.target" ];
      script = ''
        ${package}/bin/smtp-forward \
            -l ${cfg.listen} -m ${cfg.mapping} -h ${cfg.hostname} -f ${cfg.from} \
         		${lib.optionalString (cfg.key != null) "-k ${cfg.key}"} \
         		${lib.optionalString (cfg.cert != null) "-c ${cfg.cert}"}
      '';
      serviceConfig = {
        User = "smtp-forward";
	      # Security extra configuration
      	AmbientCapabilities = [ "CAP_NET_BIND_SERVICE" ];
      	CapabilityBoundingSet = [ "CAP_NET_BIND_SERVICE" ];
      	NoNewPrivileges = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        PrivateTmp = true;
        ProtectUsers = true;
        ProtectKernelLogs = true;
        PrivateDevices = true;
        ProtectHostname = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = [ "AF_INET" "AF_INET6" ];
        RestrictNamespaces = true;
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        PrivateMounts = true;
        SystemCallArchitectures = "native";
        ProtectClock = true;
        SystemCallFilter= [ "~@mount" "~@reboot" "~@swap" "~@module" "~@debug" "~@cpu-emulation" "~@obsolete" ];
      };
    };

    users.users.smtp-forward = {
      description = "smtp-forward user";
      group = "nogroup";
      extraGroups = [ "keys" ];
      uid = config.ids.uids.firebird;
    };
  };
}
