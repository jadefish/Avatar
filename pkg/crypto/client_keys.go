package crypto

import (
	"errors"

	"github.com/jadefish/avatar"
)

// these keys are from SphereServer (see
// https://github.com/Sphereserver/Source/blob/b82fba7/src/sphereCrypt.ini).
// many thanks.

// keys contains client encryption keys.
// The fourth version field ("Revision") is not used.
var keys = map[avatar.Version]avatar.KeyPair{
	{7, 0, 87, 0}: {0x0AA317E7F, 0x03CA0828D},
	{7, 0, 86, 0}: {0x0AAE5227F, 0x03D7A689D},
	{7, 0, 85, 0}: {0x0AAC95E7F, 0x03D33D2AD},
	{7, 0, 84, 0}: {0x0AAAD127F, 0x03DC480BD},
	{7, 0, 83, 0}: {0x0AA81DE7F, 0x03D8E72CD},
	{7, 0, 82, 0}: {0x0AB14227F, 0x03E5728DD},
	{7, 0, 81, 0}: {0x0AB38FE7F, 0x03E18E2ED},
	{7, 0, 80, 0}: {0x0AB3B527F, 0x03E21A0FD},
	{7, 0, 79, 0}: {0x0AB543E7F, 0x03EEB630D},
	{7, 0, 78, 0}: {0x0AB41A27F, 0x03EAC291D},
	{7, 0, 77, 0}: {0x0ABAD1E7F, 0x03F65F32D},
	{7, 0, 76, 0}: {0x0AB8A127F, 0x03F1E813D},
	{7, 0, 75, 0}: {0x0ABE79E7F, 0x03FD0534D},
	{7, 0, 74, 0}: {0x0ABCCA27F, 0x03F89695D},
	{7, 0, 73, 0}: {0x0A5D1BE7F, 0x02042036D},
	{7, 0, 72, 0}: {0x0A5C7527F, 0x0207B217D},
	{7, 0, 71, 0}: {0x0A5FDFE7F, 0x0203CC38D},
	{7, 0, 70, 0}: {0x0A598227F, 0x020E5E99D},
	{7, 0, 69, 0}: {0x0A586DE7F, 0x020AE93AD},
	{7, 0, 68, 0}: {0x0A57D127F, 0x0215781BD},
	{7, 0, 67, 0}: {0x0A5535E7F, 0x02118B3CD},
	{7, 0, 66, 0}: {0x0A521227F, 0x021C1A9DD},
	{7, 0, 65, 0}: {0x0A50F7E7F, 0x0218AA3ED},
	{7, 0, 64, 0}: {0x0A515527F, 0x021B3A1FD},
	{7, 0, 63, 0}: {0x0A484BE7F, 0x02244A40D},
	{7, 0, 62, 0}: {0x0A4A6A27F, 0x0221DAE1D},
	{7, 0, 61, 0}: {0x0A4D89E7F, 0x022D6B42D},
	{7, 0, 60, 0}: {0x0A4E2127F, 0x022AF863D},
	{7, 0, 59, 0}: {0x0A40D1E7F, 0x02360944D},
	{7, 0, 58, 0}: {0x0A41FA27F, 0x02339EE5D},
	{7, 0, 57, 0}: {0x0A47E3E7F, 0x023F2C46D},
	{7, 0, 56, 0}: {0x0A469527F, 0x023CB267D},
	{7, 0, 55, 0}: {0x0A4527E7F, 0x0238C048D},
	{7, 0, 54, 0}: {0x0A7BB227F, 0x024556E9D},
	{7, 0, 53, 0}: {0x0A7945E7F, 0x0241E54AD},
	{7, 0, 52, 0}: {0x0A715127F, 0x024E686BD},
	{7, 0, 51, 0}: {0x0A736DE7F, 0x024AFF4CD},
	{7, 0, 50, 0}: {0x0A7D6227F, 0x025702EDD},
	{7, 0, 49, 0}: {0x0A7F7FE7F, 0x0253964ED},
	{7, 0, 48, 0}: {0x0A7F7527F, 0x02501A6FD},
	{7, 0, 47, 0}: {0x0A79B3E7F, 0x025CAE50D},
	{7, 0, 46, 0}: {0x0A7B3A27F, 0x025932F1D},
	{7, 0, 45, 0}: {0x0A66A1E7F, 0x02644752D},
	{7, 0, 44, 0}: {0x0A652127F, 0x0263C873D},
	{7, 0, 43, 0}: {0x0A63A9E7F, 0x026F5D54D},
	{7, 0, 42, 0}: {0x0A612A27F, 0x026AE6F5D},
	{7, 0, 41, 0}: {0x0A6EABE7F, 0x02766856D},
	{7, 0, 40, 0}: {0x0A6F3527F, 0x0275F277D},
	{7, 0, 39, 0}: {0x0A6DAFE7F, 0x02710458D},
	{7, 0, 38, 0}: {0x0A6B2227F, 0x027C8EF9D},
	{7, 0, 37, 0}: {0x0A695DE7F, 0x0278115AD},
	{7, 0, 36, 0}: {0x0A115127F, 0x0287987BD},
	{7, 0, 35, 0}: {0x0A1345E7F, 0x0283235CD},
	{7, 0, 34, 0}: {0x0A157227F, 0x028EAAFDD},
	{7, 0, 33, 0}: {0x0A1767E7F, 0x028A325ED},
	{7, 0, 32, 0}: {0x0A169527F, 0x0289BA7FD},
	{7, 0, 31, 0}: {0x0A197BE7F, 0x0295C260D},
	{7, 0, 30, 0}: {0x0A1BCA27F, 0x02904AC1D},
	{7, 0, 29, 0}: {0x0A1D59E7F, 0x029CD362D},
	{7, 0, 28, 0}: {0x0A1EA127F, 0x029B5843D},
	{7, 0, 27, 0}: {0x0A0081E7F, 0x02A7E164D},
	{7, 0, 26, 0}: {0x0A019A27F, 0x02A26EC5D},
	{7, 0, 25, 0}: {0x0A07F3E7F, 0x02AEF466D},
	{7, 0, 24, 0}: {0x0A065527F, 0x02AD7247D},
	{7, 0, 23, 0}: {0x0A0437E7F, 0x02A9F868D},
	{7, 0, 22, 0}: {0x0A0A1227F, 0x02B406C9D},
	{7, 0, 21, 0}: {0x0A0875E7F, 0x02B08D6AD},
	{7, 0, 20, 0}: {0x0A0FD127F, 0x02BF084BD},
	{7, 0, 19, 0}: {0x0A0DBDE7F, 0x02BB976CD},
	{7, 0, 18, 0}: {0x0A328227F, 0x02C612CDD},
	{7, 0, 17, 0}: {0x0A30EFE7F, 0x02C29E6ED},
	{7, 0, 16, 0}: {0x0A313527F, 0x02C11A4FD},
	{7, 0, 15, 0}: {0x0A3723E7F, 0x02CDA670D},
	{7, 0, 14, 0}: {0x0A35DA27F, 0x02C822D1D},
	{7, 0, 13, 0}: {0x0A3B71E7F, 0x02D4AF72D},
	{7, 0, 12, 0}: {0x0A38A127F, 0x02D32853D},
	{7, 0, 11, 0}: {0x0A3ED9E7F, 0x02DFB574D},
	{7, 0, 10, 0}: {0x0A3C0A27F, 0x02DA36D5D},
	{7, 0, 9, 0}:  {0x0A223BE7F, 0x02E6B076D},
	{7, 0, 8, 0}:  {0x0A23F527F, 0x02E53257D},
	{7, 0, 7, 0}:  {0x0A21BFE7F, 0x02E1BC78D},
	{7, 0, 6, 0}:  {0x0A274227F, 0x02EC3ED9D},
	{7, 0, 5, 0}:  {0x0A250DE7F, 0x02E8B97AD},
	{7, 0, 4, 0}:  {0x0A2AD127F, 0x02F7385BD},
	{7, 0, 3, 0}:  {0x0A2895E7F, 0x02F3BB7CD},
	{7, 0, 2, 0}:  {0x0A2E5227F, 0x02FE3ADDD},
	{7, 0, 1, 0}:  {0x0A2C17E7F, 0x02FABA7ED},
	{7, 0, 0, 0}:  {0x0A2DD527F, 0x02F93A5FD},
	{6, 0, 14, 0}: {0x0A31DA27F, 0x02C022D1D},
	{6, 0, 13, 0}: {0x0A3F71E7F, 0x02DCAF72D},
	{6, 0, 12, 0}: {0x0A3CA127F, 0x02DB2853D},
	{6, 0, 11, 0}: {0x0A3AD9E7F, 0x02D7B574D},
	{6, 0, 10, 0}: {0x0A380A27F, 0x02D236D5D},
	{6, 0, 9, 0}:  {0x0A263BE7F, 0x02EEB076D},
	{6, 0, 8, 0}:  {0x0A27F527F, 0x02ED3257D},
	{6, 0, 7, 0}:  {0x0A25BFE7F, 0x02E9BC78D},
	{6, 0, 6, 0}:  {0x0A234227F, 0x02E43ED9D},
	{6, 0, 5, 0}:  {0x0A210DE7F, 0x02E0B97AD},
}

var emptyPair = &avatar.KeyPair{Lo: 0, Hi: 0}

// GetKeyPair returns a pair of client encryption keys for the provided client
// version. If no key pair exists for the provided version, an empty key pair
// and an "unsupported version" error are returned. The empty key pair should
// not be used for encryption purposes.
func GetKeyPair(version *avatar.Version) (avatar.KeyPair, error) {
	// zero out the Revision field since it is irrelevant to client keys:
	v := avatar.Version{version.Major, version.Minor, version.Patch, 0}

	if pair, ok := keys[v]; ok {
		return pair, nil
	}

	return *emptyPair, errors.New("unsupported version")
}
