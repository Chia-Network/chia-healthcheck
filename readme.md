# Chia Healthcheck

Chia Healthcheck is an application that is intended to run alongside a chia installation and return a simple healthy or unhealthy response for supported chia services.

## Installation

Download the correct executable file from the release page and run. If you are on debian/ubuntu, you can install using the apt repo, documented below.

### Apt Repo Installation

#### Set up the repository

1. Update the `apt` package index and install packages to allow apt to use a repository over HTTPS:

```shell
sudo apt-get update

sudo apt-get install ca-certificates curl gnupg
```

2. Add Chia's official GPG Key:

```shell
curl -sL https://repo.chia.net/FD39E6D3.pubkey.asc | sudo gpg --dearmor -o /usr/share/keyrings/chia.gpg
```

3. Use the following command to set up the stable repository.

```shell
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/chia.gpg] https://repo.chia.net/chia-healthcheck/debian/ stable main" | sudo tee /etc/apt/sources.list.d/chia-healthcheck.list > /dev/null
```

#### Install Chia Healthcheck

1. Update the apt package index and install the latest version of Chia Healthcheck

```shell
sudo apt-get update

sudo apt-get install chia-healthcheck
```

## Usage

First, install [chia-blockchain](https://github.com/Chia-Network/chia-blockchain). Chia healthcheck expects to be run on the same machine as the chia blockchain installation, and will use either the default chia config (`~/.chia/mainnet/`) or else the config located at `CHIA_ROOT`, if the environment variable is set.

`chia-healthcheck serve` will start the healthcheck service on the default port of `9950`.

You can check the status of the full node at `<hostname>:9950/full_node`. A response code `200` indicates the full node is receiving new blocks, while a response code of `500` would indicate that a new block has not been received within the healthcheck interval (5 minutes by default).

### Configuration

Configuration options can be passed using command line flags, environment variables, or a configuration file, except for `--config`, which is a CLI flag only. For a complete listing of options, run `chia-healthcheck --help`.

To set a config value as an environment variable, prefix the name with `CHIA_HEALTHCHECK_`, convert all letters to uppercase, and replace any dashes with underscores (`healthcheck-port` becomes `CHIA_HEALTHCHECK_HEALTHCHECK_PORT`).

To use a config file, create a new yaml file and place any configuration options you want to specify in the file. The config file will be loaded by default from `~/.chia-healthcheck.yaml`, but the location can be overridden with the `--config` flag.

```yaml
healthcheck-port: 9950
```

## Healthcheck Endpoints

This is the comprehensive list of endpoints currently available for Chia healthchecking purposes by this service.

* `/full_node` - Checks that the local full_node sync height is increasing.
* `/full_node/startup` - Checks that the local full_node sync height is increasing.
* `/full_node/liveness` - Checks that the local full_node sync height is increasing.
* `/full_node/readiness` - Checks that the local full_node is synced to the full chain.
* `/full_node/ports` - Checks that the full_node peer and RPC ports are bound.
* `/seeder` - Checks the local seeder and ensures the resolver responds with at least one peer IP address.
* `/seeder/readiness` - Checks the local seeder and ensures the resolver responds.
* `/timelord` - Checks the local timelord and ensures it is finishing proofs of time.
* `/timelord/readiness` - Checks the local timelord and ensures it is finishing proofs of time.

Other Chia components and endpoints may be added to this list over time.
