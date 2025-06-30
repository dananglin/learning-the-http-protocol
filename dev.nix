let
  commit_ref = "a676066377a2fe7457369dd37c31fd2263b662f4";
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
