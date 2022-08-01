package db_test

import (
	"testing"

	"github.com/jansemmelink/don8/db"
)

func Test1(t *testing.T) {
	g1, err := db.AddGroup(db.Group{
		Title:       "AHS/AHMP",
		Description: "Afrikaans Hoër Seuns/Meisies Skole in Pretoria",
	})
	if err != nil {
		t.Fatalf("failed: %+v", err)
	}
	t.Logf("g1: %+v", g1)

	g2, err := db.AddGroup(db.Group{
		ParentGroupID: g1.ID,
		ParentTitle:   g1.Title,
		Title:         "Wildsfees 2022",
		Description:   "Jaarlikse Wildsfees op 13 Aug 2022 vir die skool se fondsinsameling.",
	})
	if err != nil {
		t.Fatalf("failed: %+v", err)
	}
	t.Logf("g2: %+v", g2)

	for _, filter := range []string{"AHS", "AHMP", "Wildsfees"} {
		groups, err := db.FindGroups(filter, 10)
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

	db.DelGroup(g2.ID)
	db.DelGroup(g1.ID)
}
