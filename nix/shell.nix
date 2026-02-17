{
  mkShell,
  go,
  gopls,
  golangci-lint,
  nixd,
}:

mkShell {
  packages = [
    go
    gopls
    golangci-lint
    nixd
  ];
  shellHook = ''
    export GOFLAGS="-buildvcs=false"
  '';
}
