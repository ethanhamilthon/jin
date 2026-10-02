package store

import (
	"slices"
	"testing"

	"jin/internal/pricing"
	"jin/internal/provider"
)

func TestModelsCacheRoundTrip(t *testing.T) {
	db, second := openTwo(t)
	if ids, _ := second.LoadModelsCache(); ids != nil {
		t.Fatalf("empty cache: %v", ids)
	}
	if err := db.SaveModelsCache([]string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveModelLevels(map[string][]string{"a": {"low", "high"}}); err != nil {
		t.Fatal(err)
	}
	ids, _ := second.LoadModelsCache()
	levels, _ := second.LoadModelLevels()
	if !slices.Equal(ids, []string{"a", "b"}) || !slices.Equal(levels["a"], []string{"low", "high"}) || levels["b"] != nil {
		t.Fatalf("ids=%v levels=%v", ids, levels)
	}
}

func TestUsageAdd(t *testing.T) {
	var u Usage
	table := pricing.Table{"m": {InputCostPerToken: 1, OutputCostPerToken: 2}}
	u.Add(provider.Usage{Known: true, Input: 10, Output: 5}, "m", table)
	u.Add(provider.Usage{Input: 99}, "m", table)
	if u.Input != 10 || u.Output != 5 || u.Context != 15 || u.Cost != 20 {
		t.Fatalf("usage = %+v", u)
	}
}
