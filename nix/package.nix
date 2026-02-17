{ version, buildGoModule }:

buildGoModule {
  pname = "llm";
  version = version;
  src = ../.;
  vendorHash = null;
  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];
  meta = {
    mainProgram = "llm";
  };
}
