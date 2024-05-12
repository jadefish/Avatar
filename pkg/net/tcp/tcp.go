package tcp

// MaxPacketSize indicates the maximum allowed length in bytes of a packet's
// data payload.
//
// Deprecated: move to user land (outside of pkg)!
const MaxPacketSize = 1 << 16 // 65 KB

// // ByteOrder specifies the default endianness for package tcp objects.
// var ByteOrder = binary.BigEndian

// contextKey is a value for use with context.WithValue. It's used as a pointer
// so it fits in an interface{} without allocation.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return k.name
}
