# Net

* Connections established via TCP
* Packets are big-endian
* Commands come in two forms:
  1. fixed-length
  2. variable-length

## Fixed-length commands

* Byte 0: ID
* Byte 1..n: data

## Variable-length commands

* Byte 0: ID
* Byte 1..2: total size (including ID, size, and data)
* Byte 3..n: data
