# IKEA DIRIGERA Client

A client library and CLI for the IKEA DIRIGERA Smart-Home-Hub. This library is an unofficial implementation
in Go based on reverse engineering and is not affiliated with IKEA of Sweden AB! Future firmware updates may
therefore lead to incompatibility.

The current version is work in progress that does not cover all functions and has not been fully tested.
**Use this library at your own risk!**

## Install and use the library

## Install and use the CLI

### CLI Commands

| Command   | Subcommand | Args   | Flags                                       | Description                                                  | Ready |
|-----------|------------|--------|---------------------------------------------|--------------------------------------------------------------|-------|
| list      | hubs       |        | --output <format>                           | Search hubs in the local network using mDNS                  | y     |
|           | devices    |        | --output <format> --context <name>          | Lists all devices in the current context                     | y     |
|           | rooms      |        | --output <format> --context <name>          | Lists all rooms in the current context                       | y     |
|           | scenes     |        | --output <format> --context <name>          | Lists all scenes in the current context                      | y     |
|           | users      |        | --output <format> --context <name>          | List all users in the current context                        | y     |
|           | contexts   |        | --output <format>                           | List all contexts available in the config                    | y     |
| set       | context    | <name> |                                             | Set the current context                                      | y     |
| authorize |            | <ip>   | --port <port> --context <name> --no-context | Creates a new Token                                          | y     |
| show      | token      |        | --context <name>                            | Shows the access token from the current or specified context | y     |
|           | user       | <id>   | --output <format> --context <name>          | Shows information about the user                             | y     |
|           | room       | <id>   | --output <format> --context <name>          | Shows information about the room                             | y     |
|           | scene      | <id>   | --output <format> --context <name>          | Shows information about the scene                            | y     |
|           | device     | <id>   | --output <format> --context <name>          | Shows information about the device                           | y     |
| delete    | context    | <name> | --context <name>                            | Deletes the context and the user                             |       |
|           | user       | <id>   | --context <name>                            | Deletes the user                                             |       |
| curl      |            | <url>  | --context <name>                            | Calls the url in the current context                         | y     | 
| listen    |            |        | --context <name>                            | Listens for events                                           | y     |
