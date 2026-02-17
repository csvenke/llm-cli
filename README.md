# llm-cli

A lightweight CLI tool for interacting with LLMs from the terminal.

## Installation

### Nix

```bash
nix profile install github:csvenke/llm-cli
```

### Binary releases

Download pre-built binaries from [GitHub Releases](https://github.com/csvenke/llm-cli/releases).

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

[MIT](LICENSE)
