{ writeShellApplication, go }:

writeShellApplication {
  name = "test";
  runtimeInputs = [ go ];
  text = ''
    go test -race ./...
  '';
}
