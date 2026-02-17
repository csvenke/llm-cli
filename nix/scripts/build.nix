{
  writeShellApplication,
  goreleaser,
  go,
}:

writeShellApplication {
  name = "build";
  runtimeInputs = [
    goreleaser
    go
  ];
  text = ''
    goreleaser release --snapshot --clean
  '';
}
