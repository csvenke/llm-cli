{
  writeShellApplication,
  goreleaser,
  go,
  svu,
  git,
}:

writeShellApplication {
  name = "release";
  runtimeInputs = [
    goreleaser
    go
    svu
    git
  ];
  text = ''
    VERSION=$(svu next)
    git tag "$VERSION"
    goreleaser release --clean
  '';
}
