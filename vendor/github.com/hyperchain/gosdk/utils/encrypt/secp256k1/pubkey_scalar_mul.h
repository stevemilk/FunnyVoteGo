//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
int secp256k1_pubkey_scalar_mul(const secp256k1_context* ctx, unsigned char *point, const unsigned char *scalar) {
    int ret = 0;
    int overflow = 0;
    secp256k1_fe feX, feY;
    secp256k1_gej res;
    secp256k1_ge ge;
    secp256k1_scalar s;
    ARG_CHECK(point != NULL);
    ARG_CHECK(scalar != NULL);
    (void)ctx;

    secp256k1_fe_set_b32(&feX, point);
    secp256k1_fe_set_b32(&feY, point+32);
    secp256k1_ge_set_xy(&ge, &feX, &feY);
    secp256k1_scalar_set_b32(&s, scalar, &overflow);
    if (overflow || secp256k1_scalar_is_zero(&s)) {
        ret = 0;
    } else {
        secp256k1_ecmult_const(&res, &ge, &s);
        secp256k1_ge_set_gej(&ge, &res);
        /* Note: can't use secp256k1_pubkey_save here because it is not constant time. */
        secp256k1_fe_normalize(&ge.x);
        secp256k1_fe_normalize(&ge.y);
        secp256k1_fe_get_b32(point, &ge.x);
        secp256k1_fe_get_b32(point+32, &ge.y);
        ret = 1;
    }
    secp256k1_scalar_clear(&s);
    return ret;
}

