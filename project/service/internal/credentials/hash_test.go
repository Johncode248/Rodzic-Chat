package credentials

import "testing"

func TestHashToken(t *testing.T) {
	a := HashToken("token-a")
	b := HashToken("token-b")
	aAgain := HashToken("token-a")

	if a == "" {
		t.Fatal("HashToken returned an empty hash")
	}

	if a == b {
		t.Fatal("different tokens hashed to the same value")
	}

	if a != aAgain {
		t.Fatal("HashToken is not deterministic for the same input")
	}
}
