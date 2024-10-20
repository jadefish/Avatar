# Avatar

Avatar is an emulator for Ultima Online: Renaissance.

## Compatibility

Avatar supports the 2D ("classic") client, version 6.0.5.0 and later.

Client encryption is required for all connections.

## Running

TODO

### Requirements

TODO

## Building

TODO

### Requirements

TODO

## Components

TODO

### Login server

The Login server handles client authentication requests. Once successfully
authenticated, the client is provided a list of available Game servers. After
selecting a Game server, the client is relayed to the selected Game server, and
the Login server terminates its connection.

### Game server

The Game server handles character and in-game requests from clients.

Connected clients have been relayed from a Login server.

TODO
