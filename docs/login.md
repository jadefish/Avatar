# Login Flow

## Intro
To differentiate between older and newer Classic (2D) clients, read the first
byte of the first received packet and the packet's length.

_Newer_ refers to clients version 6.0.5.0 or later.

* byte `0` == `0xEF`, length == 21: newer Classic client
* byte `0` != `0xEF`, length == 4: older Classic client

## Older Classic (2D) clients
1. Client sends packet of length 4.
    * `[0, 3]`: seed (typically client's local IPv4 address)

## Newer Classic (2D) clients
1. Client sends packet of length 21 beginning with 0xEF.
    * `[0]`: `0xEF`
    * `[1, 4]`: seed (typically client's local IPv4 address)
    * `[5, 8]`: client version, major
    * `[9, 12]`: client version, minor
    * `[13, 16]`: client version, revision
    * `[17, 20]`: client version, prototype
