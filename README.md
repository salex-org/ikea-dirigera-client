# IKEA DIRIGERA Client

A client library and CLI for the IKEA DIRIGERA Smart-Home-Hub. This library is an unofficial implementation
in Go based on reverse engineering and is not affiliated with IKEA of Sweden AB! Future firmware updates may
therefore lead to incompatibility.

The current version is work in progress that does not cover all functions and has not been fully tested.
**Use this library at your own risk!**

## CLI Commands

| Command   | Subcommand | Parameter                                  | Description                                     |
|-----------|------------|--------------------------------------------|-------------------------------------------------|
| list      | hubs       |                                            | Search hubs in the local network using mDNS     |
|           | devices    |                                            | Lists all devices in the current context        |
|           | rooms      |                                            | Lists all rooms in the current context          |
|           | scenes     |                                            | Lists all scenes in the current context         |
|           | users      |                                            | List all users in the current context           |
|           | contexts   |                                            | List all contexts available in the config       |
| set       | context    | <name>                                     | Set the current context                         |
| authorize |            | <ip> --port <port> --create-context <name> | Creates a new Token                             |
| show      | token      |                                            | Shows the access token from the current context |
|           | context    |                                            | Shows information about the current context     |
| curl      |            | <url>                                      | calls the url in the current context            |
