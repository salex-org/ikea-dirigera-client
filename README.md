# IKEA DIRIGERA Client

A client library and CLI for the IKEA DIRIGERA Smart-Home-Hub. This library is an unofficial implementation
in Go based on reverse engineering and is not affiliated with IKEA of Sweden AB! Future firmware updates may
therefore lead to incompatibility.

The current version is work in progress that does not cover all functions and has not been fully tested.
**Use this library at your own risk!**

## Install and use the library

Tbd.

## Install and use the CLI

If not already done add the salex-org homebrew-tap:

```shell
brew tap salex-org/homebrew-tap
```

Install the CLI using homebrew:

```shell
brew install salex-org/tap/ikea-dirigera-cli
```

Remove the Gatekeeper-Flag to trust the binary:

```shell
xattr -d com.apple.quarantine /opt/homebrew/bin/ikea
```

Scan for IKEA DIRIGERA Hubs available on your local network:

```shell
ikea list hubs
```

Authorize a user in your IKEA DIRIGERA Hub and create a context in the CLI (replace `192.168.1.1` with the IP of your hub):

```shell
ikea authorize 192.168.1.1
```

You have to press the button on the backside of your hub to authorize the user!
Now you can use the CLI. Get information about the available commands:

```shell
ikea --help
```

