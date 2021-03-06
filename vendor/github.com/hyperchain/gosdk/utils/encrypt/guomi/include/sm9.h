
#ifndef HEADER_SM9_H
#define HEADER_SM9_H

#include <openssl/bn.h>
#include <openssl/ec.h>
#include <openssl/evp.h>
#include <openssl/asn1.h>


#ifdef __cplusplus
extern "C" {
#endif

typedef struct SM9PublicParameters_st SM9PublicParameters;
typedef struct SM9MasterSecret_st SM9MasterSecret;
typedef struct SM9PrivateKey_st SM9PrivateKey;
typedef struct SM9Ciphertext_st SM9Ciphertext;
typedef struct SM9Signature_st SM9Signature;

int SM9_setup_by_pairing_name(int nid,
	SM9PublicParameters **mpk,
	SM9MasterSecret **msk);

SM9PrivateKey *SM9_extract_private_key(SM9PublicParameters *mpk,
	SM9MasterSecret *msk,
	const char *id, size_t idlen);

typedef struct {
	const EVP_MD *kdf_md;
	const EVP_CIPHER *enc_cipher;
	const EVP_CIPHER *cmac_cipher;
	const EVP_CIPHER *cbcmac_cipher;
	const EVP_MD *hmac_md;
} SM9EncParameters;

SM9Ciphertext *SM9_do_encrypt(SM9PublicParameters *mpk,
	const SM9EncParameters *encparams,
	const unsigned char *in, size_t inlen,
	const char *id, size_t idlen);

int SM9_do_decrypt(SM9PublicParameters *mpk,
	const SM9EncParameters *encparams,
	const SM9Ciphertext *in,
	unsigned char *out, size_t *outlen,
	SM9PrivateKey *sk,
	const char *id, size_t idlen);

int SM9_encrypt(SM9PublicParameters *mpk,
	const SM9EncParameters *encparams,
	const unsigned char *in, size_t inlen,
	unsigned char *out, size_t *outlen,
	const char *id, size_t idlen);

int SM9_decrypt(SM9PublicParameters *mpk,
	const SM9EncParameters *encparams,
	const unsigned char *in, size_t inlen,
	unsigned char *out, size_t *outlen,
	SM9PrivateKey *sk,
	const char *id, size_t idlen);

int SM9_encrypt_with_recommended(SM9PublicParameters *mpk,
	const unsigned char *in, size_t inlen,
	unsigned char *out, size_t *outlen,
	const char *id, size_t idlen);

int SM9_decrypt_with_recommended(SM9PublicParameters *mpk,
	const unsigned char *in, size_t inlen,
	unsigned char *out, size_t *outlen,
	SM9PrivateKey *sk,
	const char *id, size_t idlen);

SM9Signature *SM9_do_sign(SM9PublicParameters *mpk,
	const unsigned char *dgst, size_t dgstlen,
	SM9PrivateKey *sk);

int SM9_do_verify(SM9PublicParameters *mpk,
	const unsigned char *dgst, size_t dgstlen,
	const SM9Signature *sig,
	const char *id, size_t idlen);

int SM9_sign(SM9PublicParameters *mpk,
	const unsigned char *dgst, size_t dgstlen,
	unsigned char *sig, size_t *siglen,
	SM9PrivateKey *sk);

int SM9_verify(SM9PublicParameters *mpk,
	const unsigned char *dgst, size_t dgstlen,
	const unsigned char *sig, size_t siglen,
	const char *id, size_t idlen);

DECLARE_ASN1_FUNCTIONS(SM9PublicParameters)
DECLARE_ASN1_FUNCTIONS(SM9MasterSecret)
DECLARE_ASN1_FUNCTIONS(SM9PrivateKey)
DECLARE_ASN1_FUNCTIONS(SM9Ciphertext)
DECLARE_ASN1_FUNCTIONS(SM9Signature)


/* BEGIN ERROR CODES */
/*
 * The following lines are auto generated by the script mkerr.pl. Any changes
 * made after this point may be overwritten when the script is next run.
 */

int ERR_load_SM9_strings(void);

/* Error codes for the SM9 functions. */

/* Function codes. */
# define SM9_F_SM9CIPHERTEXT_CHECK                        100
# define SM9_F_SM9ENCPARAMETERS_DECRYPT                   101
# define SM9_F_SM9ENCPARAMETERS_ENCRYPT                   102
# define SM9_F_SM9ENCPARAMETERS_GENERATE_MAC              103
# define SM9_F_SM9ENCPARAMETERS_GET_KEY_LENGTH            104
# define SM9_F_SM9PUBLICPARAMETERS_GET_POINT_SIZE         105
# define SM9_F_SM9_DECRYPT                                106
# define SM9_F_SM9_DO_DECRYPT                             107
# define SM9_F_SM9_DO_ENCRYPT                             108
# define SM9_F_SM9_DO_SIGN                                109
# define SM9_F_SM9_DO_SIGN_TYPE1CURVE                     110
# define SM9_F_SM9_DO_VERIFY                              111
# define SM9_F_SM9_DO_VERIFY_TYPE1CURVE                   112
# define SM9_F_SM9_ENCRYPT                                113
# define SM9_F_SM9_EXTRACT_PRIVATE_KEY                    114
# define SM9_F_SM9_SETUP_TYPE1CURVE                       115
# define SM9_F_SM9_SIGN                                   116
# define SM9_F_SM9_UNWRAP_KEY                             117
# define SM9_F_SM9_VERIFY                                 118
# define SM9_F_SM9_WRAP_KEY                               119

/* Reason codes. */
# define SM9_R_BUFFER_TOO_SMALL                           100
# define SM9_R_COMPUTE_PAIRING_FAILURE                    101
# define SM9_R_GENERATE_MAC_FAILURE                       102
# define SM9_R_HASH_FAILURE                               103
# define SM9_R_INVALID_CIPHERTEXT                         104
# define SM9_R_INVALID_CURVE                              105
# define SM9_R_INVALID_DIGEST                             106
# define SM9_R_INVALID_DIGEST_LENGTH                      107
# define SM9_R_INVALID_ENCPARAMETERS                      108
# define SM9_R_INVALID_ID                                 109
# define SM9_R_INVALID_ID_LENGTH                          110
# define SM9_R_INVALID_INPUT                              111
# define SM9_R_INVALID_KEY_LENGTH                         112
# define SM9_R_INVALID_MD                                 113
# define SM9_R_INVALID_PARAMETER                          114
# define SM9_R_INVALID_SIGNATURE                          115
# define SM9_R_INVALID_TYPE1CURVE                         116
# define SM9_R_KDF_FAILURE                                117
# define SM9_R_NOT_NAMED_CURVE                            118
# define SM9_R_PARSE_PAIRING                              119
# define SM9_R_ZERO_ID                                    120

# ifdef  __cplusplus
}
# endif
#endif
