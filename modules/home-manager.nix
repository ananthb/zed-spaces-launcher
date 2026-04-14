flake:
{ config, lib, pkgs, ... }:

let
  cfg = config.programs.codespace-zed;

  repoName = repository: builtins.elemAt (lib.splitString "/" repository) 1;

  targetModule = { config, lib, ... }: {
    options = {
      repository = lib.mkOption {
        type = lib.types.str;
        description = "GitHub repository in owner/repo form.";
      };

      branch = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Preferred branch when creating or matching a codespace.";
      };

      displayName = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Exact display name to disambiguate matches.";
      };

      codespaceName = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Exact codespace name for strict reuse.";
      };

      workspacePath = lib.mkOption {
        type = lib.types.str;
        default = "/workspaces/${repoName config.repository}";
        description = "Remote folder Zed should open. Defaults to /workspaces/<repo-name>.";
      };

      machine = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Machine type forwarded to gh codespace create.";
      };

      location = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Location forwarded to gh codespace create.";
      };

      devcontainerPath = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Dev container config path forwarded to gh codespace create.";
      };

      idleTimeout = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Idle timeout forwarded to gh codespace create.";
      };

      retentionPeriod = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Retention period forwarded to gh codespace create.";
      };

      uploadBinaryOverSsh = lib.mkOption {
        type = lib.types.nullOr lib.types.bool;
        default = null;
        description = "Set Zed's upload_binary_over_ssh for this host.";
      };

      zedNickname = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Friendly name shown in Zed's remote project list.";
      };

      autoStop = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Auto-stop codespace after idle duration (e.g. 30m).";
      };

      preWarm = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Time-of-day to pre-warm codespace (e.g. 08:00).";
      };
    };
  };

  filterNulls = lib.filterAttrs (_: v: v != null);

  configJSON = builtins.toJSON (filterNulls {
    defaultTarget = cfg.defaultTarget;
    targets = lib.mapAttrs (_: target: filterNulls {
      inherit (target)
        repository branch displayName codespaceName workspacePath
        machine location devcontainerPath idleTimeout retentionPeriod
        uploadBinaryOverSsh zedNickname autoStop preWarm;
    }) cfg.targets;
    daemon = lib.optionalAttrs cfg.daemon.enable (filterNulls {
      hotkey = cfg.daemon.hotkey;
      hotkeyAction = cfg.daemon.hotkeyAction;
      terminal = cfg.daemon.terminal;
      pollInterval = cfg.daemon.pollInterval;
    });
  });

  configFile = pkgs.writeText "codespace-zed-config.json" configJSON;

  wrappedPackage = pkgs.symlinkJoin {
    name = "codespace-zed-wrapped";
    paths = [ cfg.package ];
    nativeBuildInputs = [ pkgs.makeWrapper ];
    postBuild = ''
      wrapProgram $out/bin/codespace-zed \
        --add-flags "--config ${configFile}"
    '';
  };
in
{
  options.programs.codespace-zed = {
    enable = lib.mkEnableOption "codespace-zed launcher";

    package = lib.mkOption {
      type = lib.types.package;
      default = flake.packages.${pkgs.stdenv.hostPlatform.system}.default;
      description = "The codespace-zed package to use.";
    };

    defaultTarget = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      description = "Default target name when none is specified on the command line.";
    };

    targets = lib.mkOption {
      type = lib.types.attrsOf (lib.types.submodule targetModule);
      default = { };
      description = "Named codespace targets.";
    };

    daemon = {
      enable = lib.mkOption {
        type = lib.types.bool;
        default = true;
        description = "Whether to enable the codespace-zed background daemon (tray, hotkey, lifecycle).";
      };

      hotkey = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = if pkgs.stdenv.isDarwin then "Cmd+Shift+S" else "Ctrl+Shift+S";
        description = "Global hotkey to open the codespace picker.";
      };

      hotkeyAction = lib.mkOption {
        type = lib.types.enum [ "picker" "previous" "default" ];
        default = "picker";
        description = ''
          What the hotkey does:
          - "picker": open the interactive repo/codespace picker
          - "previous": launch the most recently used target
          - "default": launch the default target from config
        '';
      };

      terminal = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Terminal application for launching the picker (null for auto-detect).";
      };

      pollInterval = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = "5m";
        description = "Interval for polling codespace states.";
      };
    };
  };

  config = lib.mkIf cfg.enable {
    home.packages = [ wrappedPackage ];

    programs.ssh.includes = [ "~/.ssh/codespaces-zed/*.conf" ];

    home.file.".ssh/codespaces-zed/.keep".text = "";

    # Copy .app bundle into ~/Applications on activation.
    # home.file with recursive=true creates per-file symlinks which macOS
    # does not recognise as a valid bundle. Instead, copy the whole .app
    # directory and strip the quarantine xattr so Gatekeeper accepts it.
    home.activation.codespace-zed-app = lib.mkIf pkgs.stdenv.isDarwin
      (lib.hm.dag.entryAfter [ "writeBoundary" ] ''
        app_src="${wrappedPackage}/Applications/Codespace Zed.app"
        app_dst="$HOME/Applications/Codespace Zed.app"
        $DRY_RUN_CMD rm -rf "$app_dst"
        $DRY_RUN_CMD cp -RL "$app_src" "$app_dst"
        $DRY_RUN_CMD chmod -R u+w "$app_dst"
        $DRY_RUN_CMD xattr -dr com.apple.quarantine "$app_dst" 2>/dev/null || true
      '');

    # macOS launchd agent for the daemon.
    launchd.agents.codespace-zed-daemon = lib.mkIf (cfg.daemon.enable && pkgs.stdenv.isDarwin) {
      enable = true;
      config = {
        # Launch from the .app bundle so macOS associates the process
        # with the bundle (dock icon, app lifecycle, permissions).
        # The .app binary only has the gh PATH wrapper — it doesn't
        # include the home-manager --config wrapper, so pass it here.
        ProgramArguments = [
          "${wrappedPackage}/Applications/Codespace Zed.app/Contents/MacOS/codespace-zed"
          "--config" "${configFile}"
          "applet"
        ];
        # Only restart on abnormal exit — lets the user quit cleanly
        # via the tray menu without launchd immediately restarting.
        KeepAlive = { SuccessfulExit = false; };
        RunAtLoad = true;
        Label = "com.codespace-zed.daemon";
        StandardOutPath = "${config.home.homeDirectory}/Library/Logs/codespace-zed.log";
        StandardErrorPath = "${config.home.homeDirectory}/Library/Logs/codespace-zed.log";
        ProcessType = "Interactive";
      };
    };

    # Linux systemd user service for the daemon.
    systemd.user.services.codespace-zed-daemon = lib.mkIf (cfg.daemon.enable && pkgs.stdenv.isLinux) {
      Unit = {
        Description = "codespace-zed background daemon";
        After = [ "graphical-session.target" ];
        PartOf = [ "graphical-session.target" ];
      };
      Service = {
        ExecStart = "${wrappedPackage}/bin/codespace-zed daemon start";
        Restart = "on-failure";
        RestartSec = 5;
      };
      Install = {
        WantedBy = [ "graphical-session.target" ];
      };
    };
  };
}
