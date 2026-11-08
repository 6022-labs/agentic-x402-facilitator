package helpers

// SplitSignature splits a 65-byte r||s||v signature, normalizing v to 27/28 as ecrecover expects.
func SplitSignature(signature []byte) (uint8, [32]byte, [32]byte) {
	var r, s [32]byte
	copy(r[:], signature[0:32])
	copy(s[:], signature[32:64])
	v := signature[64]
	if v < 27 {
		v += 27
	}
	return v, r, s
}
