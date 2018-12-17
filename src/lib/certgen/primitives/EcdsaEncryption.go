package primitives

// EcdsaEncryption .
type EcdsaEncryption struct {
	name string
	id   string
}

// NewEcdsaEncrypto .
func NewEcdsaEncrypto(name string) *EcdsaEncryption {
	ee := &EcdsaEncryption{name: name}
	return ee
}

// Sign .
func (ee *EcdsaEncryption) Sign(payload []byte, pri interface{}) ([]byte, error) {
	//log.Notice("payload",payload,"pri",pri)

	sign, err := ECDSASign(pri, payload)

	if err != nil {
		return nil, err
	}

	return sign, err

}

// VerifySign .
func (ee *EcdsaEncryption) VerifySign(verKey interface{}, msg, signature []byte) (bool, error) {

	result, err := ECDSAVerify(verKey, msg, signature)

	if err != nil {
		return false, err
	}
	return result, nil
}
