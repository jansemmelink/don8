package db_test

import (
	"testing"

	"github.com/jansemmelink/don8/db"
)

func TestGroups(t *testing.T) {
	u, err := db.AddUser(db.User{Name: "A", Phone: "0821111111", Email: "a@b.c"})
	if err != nil {
		t.Fatalf("failed to create user for group member")
	}
	defer func() {
		db.DelUser(u.ID)
	}()

	desc1 := "Afrikaans Hoër Seuns/Meisies Skole in Pretoria"
	g1, err := db.AddGroup(u, db.NewGroup{
		Title:       "AHS/AHMP",
		Description: &desc1,
	})
	if err != nil {
		t.Fatalf("failed: %+v", err)
	}
	defer func() {
		db.DelGroup(g1.ID)
	}()
	t.Logf("g1: %+v", g1)

	desc2 := "Jaarlikse Wildsfees op 13 Aug 2022 vir die skool se fondsinsameling."
	g2, err := db.AddGroup(u, db.NewGroup{
		ParentGroupID: g1.ID,
		Title:         "Wildsfees 2022",
		Description:   &desc2,
	})
	if err != nil {
		t.Fatalf("failed: %+v", err)
	}
	t.Logf("g2: %+v", g2)
	defer func() {
		db.DelGroup(g2.ID)
	}()

	for _, filter := range []string{"AHS", "AHMP", "Wildsfees"} {
		groups, err := db.MyGroups(u, filter, nil, nil)
		if err != nil {
			t.Fatalf("failed to find %s: %+v", filter, err)
		}
		t.Logf("Found %d %s groups", len(groups), filter)
		for _, g := range groups {
			t.Logf("  %s: %+v", filter, g)
		}
	}

	if g, err := db.GetGroup(g1.ID); err != nil {
		t.Fatalf("failed to get g1: %+v", err)
	} else {
		if err := g1.Compare(g); err != nil {
			t.Fatalf("g1!=g: %+v", err)
		}
	}
	if g, err := db.GetGroup(g2.ID); err != nil {
		t.Fatalf("failed to get g2: %+v", err)
	} else {
		if err := g2.Compare(g); err != nil {
			t.Fatalf("g2!=g: %+v", err)
		}
	}
}
