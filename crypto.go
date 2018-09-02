package avatar

type ClientKey uint32

type Crypto struct {
	Seed     uint32
	MasterLo ClientKey
	MasterHi ClientKey
	MaskLo   uint32
	MaskHi   uint32
}
