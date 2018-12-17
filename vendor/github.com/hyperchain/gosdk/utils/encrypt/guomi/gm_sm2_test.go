package guomi

import (
	"github.com/hyperchain/gosdk/common"
	"testing"
)

/*
pri=bf873403306b6691f3fddd384e4c45e2acebefaaa7403b1e9bc8f5acd00c6bd8
pub=0x02f9c11f5a8e899aa857ab9551900dec62687abe63828761b1bd18332d3f9cb73c

supPrivateKey1 = "f0f562eb8b76e6399c601e76431ccbb31e289980effd20d9f2236350005adb1f"
 supPublicKey1 = "0x0238e60dc6395e50119342a7829f266deb22efb290789506a54f5b6bfcca2f88f2"

 pri=362bf3fbd7308925f075bf3ab5396c931d9f89ec68b3585e3f6541015b33706a
 pub=0x0329ea23368f4c13815781ffd6259f16adf00cb5712014f7ebdf9faf7acd1bf9d4
*/

func TestUncompressedPubkey(t *testing.T) {
	pubX, _ := UncompressedPubkeyOpenssl("0x0329ea23368f4c13815781ffd6259f16adf00cb5712014f7ebdf9faf7acd1bf9d4")
	t.Log(pubX)

	pri, _ := GetPriKeyFromHex(common.Hex2Bytes("362bf3fbd7308925f075bf3ab5396c931d9f89ec68b3585e3f6541015b33706a"))
	pub2 := GetPubKeyFromPri(pri)
	t.Log(pub2[33:])
}
