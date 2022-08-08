package db_test

import (
	"testing"

	"github.com/jansemmelink/don8/db"
)

func TestUsers(t *testing.T) {
	//todo: nead to test in a clean db! consider using sqlite for testing
	u1, err := db.AddUser(db.User{
		Name:  "Jan Semmelink",
		Phone: "0821111111",
		Email: "jan.1111111@gmail.com",
	})
	if err != nil {
		t.Fatalf("failed: %+v", err)
	}
	t.Logf("u1: %+v", u1)
	defer func() {
		db.DelUser(u1.ID)
	}()

	u2, err := db.AddUser(db.User{
		Name:  "Koos",
		Phone: "0821234567",
		Email: "Koos@gmail.com",
	})
	if err != nil {
		t.Fatalf("failed: %+v", err)
	}
	t.Logf("u2: %+v", u2)
	defer func() {
		db.DelUser(u2.ID)
	}()

	for _, filter := range []string{"Jan", "Koos", "082"} {
		users, err := db.FindUsers(filter, 10)
		if err != nil {
			t.Fatalf("failed to find %s: %+v", filter, err)
		}
		t.Logf("Found %d %s users", len(users), filter)
		for _, u := range users {
			t.Logf("  %s: %+v", filter, u)
		}
	}

	if u, err := db.GetUser(u1.ID); err != nil {
		t.Fatalf("failed to get u1: %+v", err)
	} else {
		if err := u1.Compare(u); err != nil {
			t.Fatalf("u1!=u: %+v", err)
		}
	}
	if u, err := db.GetUser(u2.ID); err != nil {
		t.Fatalf("failed to get u2: %+v", err)
	} else {
		if err := u2.Compare(u); err != nil {
			t.Fatalf("u2!=u: %+v", err)
		}
	}
}
