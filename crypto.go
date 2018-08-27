package avatar

type ClientKey uint32

type Crypto struct {
	MasterLo ClientKey
	MasterHi ClientKey
	MaskLo   uint32
	MaskHi   uint32
}
