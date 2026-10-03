package pricing

import (
	"math"
	"testing"
)

func TestCostWithCacheWrite(t *testing.T) {
	entry := Entry{InputCostPerToken: 2, CacheReadCostPerToken: 1, CacheWriteCostPerToken: 3, OutputCostPerToken: 10}
	got := entry.CostWithCacheWrite(100, 40, 20, 5)
	if want := 40.0*2 + 20*3 + 40*1 + 5*10; math.Abs(got-want) > 1e-9 {
		t.Errorf("cost = %v, want %v", got, want)
	}
	if got, want := entry.CostWithCacheWrite(10, 40, 50, 0), 10.0; math.Abs(got-want) > 1e-9 {
		t.Errorf("clamped cost = %v, want %v", got, want)
	}
	plain := Entry{InputCostPerToken: 2}
	if got := plain.CostWithCacheWrite(10, 0, 10, 0); got != 20 {
		t.Errorf("no write rate = %v", got)
	}
}

func TestParseCacheWriteRates(t *testing.T) {
	router, err := parseOpenRouter([]byte(`{"data":[{"id":"a/m","pricing":{"input_cache_write":"0.5"}}]}`))
	if err != nil || router["a/m"].CacheWriteCostPerToken != 0.5 {
		t.Errorf("openrouter = %+v, %v", router["a/m"], err)
	}
	lite, err := parseLiteLLM([]byte(`{"m":{"cache_creation_input_token_cost":0.25}}`))
	if err != nil || lite["m"].CacheWriteCostPerToken != 0.25 {
		t.Errorf("litellm = %+v, %v", lite["m"], err)
	}
	if _, ok := (Table{"gpt-6": {}}).Lookup("gpt-6-xhigh"); !ok {
		t.Error("xhigh suffix not stripped")
	}
}
