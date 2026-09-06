package service

import (
	"math"
	"testing"
	"github.com/xrekson/game-backend/services/economy/model"
)

func TestUpgradeCostScaling(t *testing.T) {
	baseCost := model.DefaultUpgrades[0].BaseCost // 15.0

	costLvl0 := baseCost * math.Pow(1.15, 0)
	if costLvl0 != 15.0 {
		t.Errorf("expected level 0 cost to be 15.0, got %f", costLvl0)
	}

	costLvl1 := baseCost * math.Pow(1.15, 1)
	expectedLvl1 := 17.25
	if math.Abs(costLvl1-expectedLvl1) > 0.01 {
		t.Errorf("expected level 1 cost to be ~17.25, got %f", costLvl1)
	}
}
