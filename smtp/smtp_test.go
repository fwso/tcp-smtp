package smtp

import "testing"

func TestAuth(t *testing.T) {
	user := "linhua"
	pass := "testpassword"
	host := "testserver:25"

	c, err := Dial(host)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	if err := c.EHLO("localhost"); err != nil {
		t.Fatalf("error: %v\n", err)
	}

	if err := c.Auth(user, pass); err != nil {
		t.Fatalf("error: %v\n", err)
	}
}
