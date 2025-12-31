#!/usr/bin/env bash

# Dummy script to simulate a client sending a partial command to a login server.
# Sends 20 out of the required 21 bytes for the Login Seed packet, then sends
# the final byte 3 seconds later.

# TODO: write a test for this case and remove this script.

(
  printf "\xEF\xC0\xA8\x00\x01\x00\x00\x00\x10\x00\x00\x00\x11\x00\x00\x00\x12\x00\x00\x00"
  sleep 3
  printf "\x13"
) | socat - TCP:localhost:7775
