package report

import (
	"fmt"
	"testing"
	"time"
)

// RSA公钥
const rsaPublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsu2yA4PcEW8huBJyhgp/
djOLglIggi5c4i95AFLpvbl4EhyazIMPYaGakpCwhjiFkYKSVkJa7PUgfVzwRamX
7ydkLqSsgC91g7sgKhGEZkMz+bGmXmWawNG7EfOxVq9p3X7tgfHQaHKE41cUp+0q
JmH4/m8jg+tYKVaKOGjIGSHXbNJ1yO36z/I3bmPD9a9EZQivrxDFXmD6TLemONKo
UOnW4x2hWTkVNJ5zjeXDuy45GgDcFxe3lfnZYuuR5iV84BLgDIaINS1++MvYLeA1
VchbrPYgj4rhw9WikRcxMBydOEUygVZhswW8xAyADO4KDTVMDY9syMb3gXDj8tDp
RwIDAQAB
-----END PUBLIC KEY-----`

func TestSend(t *testing.T) {
	reportHead := ReportHead{UserId: 0}
	nowMillis := time.Now().UnixMilli()
	reportPayload := ReportPayload{
		{
			OpRes:     1,
			OpObjType: 31,
			OpObjValue: map[string]interface{}{
				"url": "www1.airdrop.signglobal.info",
			},
			UserTimestamp: nowMillis,
			Timestamp:     nowMillis,
		},
		{
			OpRes:     1,
			OpObjType: 31,
			OpObjValue: map[string]interface{}{
				"url": "www2.airdrop.signglobal.info",
			},
			UserTimestamp: nowMillis,
			Timestamp:     nowMillis,
		},
		{
			OpRes:     1,
			OpObjType: 31,
			OpObjValue: map[string]interface{}{
				"url": "www3.airdrop.signglobal.info",
			},
			UserTimestamp: nowMillis,
			Timestamp:     nowMillis,
		},
	}
	err := send("https://sys-test.adspower.net", "/conf", rsaPublicKeyPEM, reportHead, reportPayload)
	if err != nil {
		fmt.Printf("Send failed: %v", err)
	}
}
