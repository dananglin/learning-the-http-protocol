let
  commit_ref = "70c74b02eac46f4e4aa071e45a6189ce0f6d9265";
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

  TMUX_SESSION = "bootdev-http-protocol";

  shellHook = ''
    export GOROOT=$( which go | xargs dirname | xargs dirname )/share/go
    tmux new-session -d -s "$TMUX_SESSION"
    exec tmux attach-session -t "$TMUX_SESSION"
  '';
}
