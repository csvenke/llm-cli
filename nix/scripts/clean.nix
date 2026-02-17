{ writeShellApplication, coreutils }:

writeShellApplication {
  name = "clean";
  runtimeInputs = [ coreutils ];
  text = ''
    rm -rf dist/
  '';
}
