#!/usr/bin/env ruby

version = ARGV[0]

if version.to_s.strip.empty?
  warn "usage: key_gen version-string"
  exit(1)
end

parts = version.split(".", 4).map(&:to_i)
version_int = (parts[0] << 24) | (parts[1] << 16) | (parts[2] << 8) | parts[3]

warn "#{version.inspect} -> #{sprintf("0x%X", version_int)} (#{version_int})"

# https://github.com/ClassicUO/ClassicUO/blob/3ad74a6/src/ClassicUO.Client/Network/Encryption/Encryption.cs#L67-L98
=begin
int a = ((int)version >> 24) & 0xFF;
int b = ((int)version >> 16) & 0xFF;
int c = ((int)version >> 8) & 0xFF;

int temp = ((((a << 9) | b) << 10) | c) ^ ((c * c) << 5);

var key2 = (uint)((temp << 4) ^ (b * b) ^ (b * 0x0B000000) ^ (c * 0x380000) ^ 0x2C13A5FD);
temp = (((((a << 9) | c) << 10) | b) * 8) ^ (c * c * 0x0c00);
var key3 = (uint)(temp ^ (b * b) ^ (b * 0x6800000) ^ (c * 0x1c0000) ^ 0x0A31D527F);
var key1 = key2 - 1;
=end

a = (version_int >> 24) & 0xFF
b = (version_int >> 16) & 0xFF
c = (version_int >> 8) & 0xFF
# the fourth version component is not used.

temp = ((((a << 9) | b) << 10) | c) ^ ((c * c) << 5)
puts temp

key2 = ((temp << 4) ^ (b * b) ^ (b * 0x0B000000) ^ (c * 0x380000) ^ 0x2C13A5FD) & 0xFFFFFFFF
puts key2

temp = (((((a << 9) | c) << 10) | b) * 8) ^ (c * c * 0x0c00)
puts temp

key3 = (temp ^ (b * b) ^ (b * 0x6800000) ^ (c * 0x1c0000) ^ 0x0A31D527F) & 0xFFFFFFFF
puts key3

key1 = (key2 - 1) & 0xFFFFFFFF
puts key1

puts [key1, key2, key3].map { |n| sprintf("0x%X", n) }.join(" ")
