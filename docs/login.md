# Login Flow

## Intro
To differentiate between older and newer Classic (2D) clients, read the first
byte of the first received packet.

_Newer_ refers to client version 6.0.5.0 or later.

* byte `0` == `0xEF`, length == 21: newer Classic client
* byte `0` != `0xEF`, length == 4: older Classic client

## Older Classic (2D) clients
1. receive 4 bytes: legacy login seed
2. (TODO: document older client login flow)

## Newer Classic (2D) clients
### Login server
1. receive 21 bytes: 0xEF Login Seed
2. receive 62 bytes: 0x80 Login Request (encrypted)
3. send n bytes: 0xA8 Game Server List
4. (optional) receive 268 bytes: 0xD9 Client Info (encrypted)
	* not variable length, but length has changed over time with newer clients
5. receive 3 bytes: 0xA0 Select Server (encrypted)
6. send 11 bytes: 0x8C Connect to Game Server
   * last 4 bytes: "new key" (uint32). if not specified, client seed from 0xEF
	 is used.
7. client now establishes a new connection to the selected game server

### Game server
1. receive 69 bytes:
	* 4 bytes: "new key" (last 4 bytes, uint32) from 0x8C Connect to Game Server
	* 65 bytes: 0x91 Game Server Login (encrypted)
		* invariant: decrypted bytes[0] === 0x91
