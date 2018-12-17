//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package secp256k1

// TODO: set USE_SCALAR_4X64 depending on platform?

/*
#cgo CFLAGS: -I./libsecp256k1
#cgo CFLAGS: -I./libsecp256k1/src/
#define USE_NUM_NONE
#define USE_FIELD_10X26
#define USE_FIELD_INV_BUILTIN
#define USE_SCALAR_8X32
#define USE_SCALAR_INV_BUILTIN
#define NDEBUG
#include "./libsecp256k1/src/secp256k1.c"
#include "./libsecp256k1/src/modules/recovery/main_impl.h"
#include "pubkey_scalar_mul.h"

typedef void (*callbackFunc) (const char* msg, void* data);
extern void secp256k1GoPanicIllegal(const char* msg, void* data);
extern void secp256k1GoPanicError(const char* msg, void* data);
*/
import "C"

import (
	"errors"
	"math/big"
	"unsafe"
)

//#define USE_FIELD_5X64

// holds ptr to secp256k1_context_struct (see secp256k1/include/secp256k1.h)
var (
	context *C.secp256k1_context
	N       *big.Int
	HalfN   *big.Int
)

func init() {
	N, _ = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	// N / 2 == 57896044618658097711785492504343953926418782139537452191302581570759080747168
	HalfN, _ = new(big.Int).SetString("7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0", 16)

	// around 20 ms on a modern CPU.
	context = C.secp256k1_context_create(3) // SECP256K1_START_SIGN | SECP256K1_START_VERIFY
	C.secp256k1_context_set_illegal_callback(context, C.callbackFunc(C.secp256k1GoPanicIllegal), nil)
	C.secp256k1_context_set_error_callback(context, C.callbackFunc(C.secp256k1GoPanicError), nil)
}

var (
	ErrInvalidMsgLen       = errors.New("invalid message length for signature recovery")
	ErrInvalidSignatureLen = errors.New("invalid signature length")
	ErrInvalidRecoveryID   = errors.New("invalid signature recovery id")
)

func Sign(msg []byte, seckey []byte) ([]byte, error) {
	msg_ptr := (*C.uchar)(unsafe.Pointer(&msg[0]))
	seckey_ptr := (*C.uchar)(unsafe.Pointer(&seckey[0]))

	sig := make([]byte, 65)
	sig_ptr := (*C.secp256k1_ecdsa_recoverable_signature)(unsafe.Pointer(&sig[0]))

	nonce, err := csprng(32)
	if err != nil {
		return nil, err
	}
	ndata_ptr := unsafe.Pointer(&nonce[0])

	noncefp_ptr := &(*C.secp256k1_nonce_function_default)

	if C.secp256k1_ec_seckey_verify(context, seckey_ptr) != C.int(1) {
		return nil, errors.New("Invalid secret key")
	}

	ret := C.secp256k1_ecdsa_sign_recoverable(
		context,
		sig_ptr,
		msg_ptr,
		seckey_ptr,
		noncefp_ptr,
		ndata_ptr,
	)

	if ret == C.int(0) {
		return Sign(msg, seckey) //invalid secret, try again
	}

	sig_serialized := make([]byte, 65)
	sig_serialized_ptr := (*C.uchar)(unsafe.Pointer(&sig_serialized[0]))
	var recid C.int

	C.secp256k1_ecdsa_recoverable_signature_serialize_compact(
		context,
		sig_serialized_ptr, // 64 byte compact signature
		&recid,
		sig_ptr, // 65 byte "recoverable" signature
	)

	sig_serialized[64] = byte(int(recid)) // add back recid to get 65 bytes sig

	return sig_serialized, nil

}

// RecoverPubkey returns the the public key of the signer.
// msg must be the 32-byte hash of the message to be signed.
// sig must be a 65-byte compact ECDSA signature containing the
// recovery id as the last element.
func RecoverPubkey(msg []byte, sig []byte) ([]byte, error) {
	if len(msg) != 32 {
		return nil, ErrInvalidMsgLen
	}
	if err := checkSignature(sig); err != nil {
		return nil, err
	}

	msg_ptr := (*C.uchar)(unsafe.Pointer(&msg[0]))
	sig_ptr := (*C.uchar)(unsafe.Pointer(&sig[0]))
	pubkey := make([]byte, 64)
	/*
		this slice is used for both the recoverable signature and the
		resulting serialized pubkey (both types in libsecp256k1 are 65
		bytes). this saves one allocation of 65 bytes, which is nice as
		pubkey recovery is one bottleneck during load in Ethereum
	*/
	bytes65 := make([]byte, 65)
	pubkey_ptr := (*C.secp256k1_pubkey)(unsafe.Pointer(&pubkey[0]))
	recoverable_sig_ptr := (*C.secp256k1_ecdsa_recoverable_signature)(unsafe.Pointer(&bytes65[0]))
	recid := C.int(sig[64])

	ret := C.secp256k1_ecdsa_recoverable_signature_parse_compact(
		context,
		recoverable_sig_ptr,
		sig_ptr,
		recid)
	if ret == C.int(0) {
		return nil, errors.New("Failed to parse signature")
	}

	ret = C.secp256k1_ecdsa_recover(
		context,
		pubkey_ptr,
		recoverable_sig_ptr,
		msg_ptr,
	)
	if ret == C.int(0) {
		return nil, errors.New("Failed to recover public key")
	}

	serialized_pubkey_ptr := (*C.uchar)(unsafe.Pointer(&bytes65[0]))
	var output_len C.size_t
	C.secp256k1_ec_pubkey_serialize( // always returns 1
		context,
		serialized_pubkey_ptr,
		&output_len,
		pubkey_ptr,
		0, // SECP256K1_EC_COMPRESSED
	)
	return bytes65, nil
}

func checkSignature(sig []byte) error {
	if len(sig) != 65 {
		return ErrInvalidSignatureLen
	}
	if sig[64] >= 4 {
		return ErrInvalidRecoveryID
	}
	return nil
}

// reads num into buf as big-endian bytes.
func readBits(buf []byte, num *big.Int) {
	const wordLen = int(unsafe.Sizeof(big.Word(0)))
	i := len(buf)
	for _, d := range num.Bits() {
		for j := 0; j < wordLen && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
}
