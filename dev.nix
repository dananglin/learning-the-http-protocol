let
  # Branch: nixos-unstable
  # Date of commit: 2025-10-31
  commit_ref = "2fb006b87f04c4d3bdf08cfdbc7fab9c13d94a15";
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/${commit_ref}";
  pkgs = import nixpkgs {
    config = { };
    overlays = [ ];
  };
in

pkgs.mkShellNoCC {
  packages = with pkgs; [
    go
    golangci-lint
    gopls
  ];

  TMUX_SESSION = "Learn the HTTP Protocol";

  shellHook = ''
    export GOROOT=$( which go | xargs dirname | xargs dirname )/share/go
    tmux new-session -d -s "$TMUX_SESSION"
    exec tmux attach-session -t "$TMUX_SESSION"
  '';
}
