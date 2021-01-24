# cast-cli (Experimental)

NOTE: CAST AI CLI is in it's early stage. Feel free to contribute, ask questions, give feedback.

## Installation

### macOS

`cast` is available via [Homebrew][], and as a downloadable binary from the [releases page][].

#### Homebrew

| Install:          | Upgrade:          |
| ----------------- | ----------------- |
| `brew install castai/tap/cast-cli` | `brew upgrade castai/tap/cast-cli` |

### Linux

`cast` is available via [Homebrew](#homebrew), and as downloadable binaries from the [releases page][].

#### Homebrew

| Install:          | Upgrade:          |
| ----------------- | ----------------- |
| `brew install castai/tap/cast-cli` | `brew upgrade castai/tap/cast-cli` |

### Windows

`cast` is available as a downloadable binary from the [releases page][].

## Getting started

After installing CLI you need to configure API access token to access CAST AI public API.

### Quick configuration

```
cast configure
```

After done configuration file is saved to file system.
	
### Configure via environment variables
It is possible to override all configuration with environment variables.

| Variable          | Description          | Default |
| ----------------- | ----------------- | ----------------- |
| CASTAI_API_TOKEN | API access token | - |
| CASTAI_API_HOSTNAME | API access token | api.cast.ai |
| CASTAI_DEBUG | Enable debug mode | false |
| CASTAI_CONFIG | Path to CLI configuration file | ~/.cast.config |

## Usage

Run `cast` without any arguments to get help. Use --help on sub commands to get more help, eg. `cast cluster --help` 

```
CAST AI Command Line Interface

Usage:
  cast [command]

Available Commands:
  cluster     Manage clusters
  completion  Generate completion script
  configure   Setup initial configuration
  credentials Manage credentials
  help        Help about any command
  region      Manage regions
  version     Print version

Flags:
  -h, --help   help for cast

Use "cast [command] --help" for more information about a command.
```

[Homebrew]: https://brew.sh
[releases page]: https://github.com/castai/cast-cli/releases/latest
