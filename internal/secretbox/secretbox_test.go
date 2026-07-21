package secretbox

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	ciphertext, nonce, err := Encrypt(key, []byte("tenant-secret"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decrypt(key, ciphertext, nonce)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "tenant-secret" {
		t.Fatalf("got %q", got)
	}
	if string(ciphertext) == "tenant-secret" {
		t.Fatal("ciphertext contains plaintext")
	}
}
