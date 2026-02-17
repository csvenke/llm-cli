{ writeShellApplication, golangci-lint }:

writeShellApplication {
  name = "lint";
  runtimeInputs = [ golangci-lint ];
  text = ''
    golangci-lint run ./...
  '';
}
