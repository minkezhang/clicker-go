package cookie_clicker

import (
	"fmt"
	"testing"
)

func TestMakeSimpleUpgrade(t *testing.T) {
	s := NewGameState()
	u := NewBuildingUpgrade(
		BUILDING_TYPE_MOUSE,
		"New Upgrade",
		100,
		2,
		0,
	)
	if (*u).GetName() != "New Upgrade" {
		t.Error(fmt.Sprintf("Expected name %s, got %s", "New Upgrade", (*u).GetName()))
	}

	if (*u).GetCost(s) != 100 {
		t.Error(fmt.Sprintf("Expected upgrade cost %e, got %e", 100, (*u).GetCost(s)))
	}

	if (*u).GetBuildingMultiplier(s) != 2 {
		t.Error(fmt.Sprintf("Expected BMul %e, got %e", 2, (*u).GetBuildingMultiplier(s)))
	}

	if !(*u).GetIsUnlocked(s) {
		t.Error("Expected upgrade is unlocked, but wasn't")
	}
}

func TestSimpleUpgradeIsUnlocked(t *testing.T) {
	s := NewGameState()
	u := NewBuildingUpgrade(
		BUILDING_TYPE_MOUSE,
		"New Upgrade",
		100,
		2,
		1,
	)

	if (*u).GetIsUnlocked(s) {
		t.Error("Expected upgrade is locked, but wasn't")
	}
}