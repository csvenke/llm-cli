# llm-cli

A lightweight CLI tool for interacting with LLMs from the terminal.

## Installation

### Binary releases

For Windows, macOS 12 or newer, or Linux, you can download a binary release [here](https://github.com/csvenke/llm-cli/releases)

### Nix

#### nix profile

```bash
nix profile add github:csvenke/llm-cli
```

#### nix flake

```nix
{
  description = "your flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    dev-cli = {
      url = "github:csvenke/llm-cli";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };
}
```

## Usage

### Generate a commit message

Stage your changes and run:

```bash
llm commit
```

To amend the previous commit:

```bash
llm commit -a
```

### Ask a question

```bash
llm ask "How do I reverse a string in Go?"
```

## Configuration

Set one of the following environment variables (checked in order):

| Variable                  | Description              |
| ------------------------- | ------------------------ |
| `OPENCODE_ZEN_API_KEY`    | OpenCode Zen API key     |
| `ANTHROPIC_API_KEY`       | Anthropic API key        |
| `OPENAI_API_KEY`          | OpenAI API key           |


## License

MIT
