package util

import "testing"


func TestDes(t *testing.T) {
	key := []byte("2fa6c1e9")
	str := "I love this beautiful world!"
	// create test data

	strEncrypted, err := DesEncrypt(str, key)
	if err != nil {
		t.Fail()
	}
	strDecrypted, err := DesDecrypt(strEncrypted, key)
	if err != nil || strDecrypted != str {
		t.Fail()
	}

}

func TestAes(t *testing.T) {
	key := "2fa6c1e9oaferqes"
	str := "I love this beautiful world!"

	strEncrypted,_ := AesEncrypt(str, key)
	strDecrypted,_ := AesDecrypt(strEncrypted, key)
	if strDecrypted != str {
		t.Fail()
	}
}