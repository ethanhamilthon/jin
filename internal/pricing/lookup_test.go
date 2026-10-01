package pricing

import (
	"math"
	"testing"
)

func TestLookup(t *testing.T) {
	table := Table{
		"gpt-5":                     {InputCostPerToken: 1},
		"anthropic/claude-sonnet-5": {InputCostPerToken: 2},
	}
	for _, model := range []string{"gpt-5", "openai/gpt-5", "gpt-5-high", "claude-sonnet-5"} {
		if _, ok := table.Lookup(model); !ok {
			t.Errorf("Lookup(%q) not found", model)
		}
	}
	if _, ok := table.Lookup("unknown"); ok {
		t.Error("Lookup(unknown) should not be found")
	}
}

func TestCost(t *testing.T) {
	entry := Entry{InputCostPerToken: 2, CacheReadCostPerToken: 1, OutputCostPerToken: 10}
	got := entry.Cost(100, 40, 5)
	if want := 60.0*2 + 40*1 + 5*10; math.Abs(got-want) > 1e-9 {
		t.Errorf("Cost = %v, want %v", got, want)
	}
}

func TestParseMaxInputTokens(t *testing.T) {
	router, err := parseOpenRouter([]byte(`{"data":[{"id":"a/m","context_length":200000,"pricing":{"prompt":"0.1"}}]}`))
	if err != nil || router["a/m"].MaxInputTokens != 200000 {
		t.Errorf("openrouter = %+v, %v", router["a/m"], err)
	}
	lite, err := parseLiteLLM([]byte(`{"m":{"max_input_tokens":128000}}`))
	if err != nil || lite["m"].MaxInputTokens != 128000 {
		t.Errorf("litellm = %+v, %v", lite["m"], err)
	}
}

func TestVisionFromCatalogues(t *testing.T) {
	router, err := parseOpenRouter([]byte(`{"data":[{"id":"a/v","architecture":{"input_modalities":["text","image"]}},{"id":"a/t","architecture":{"input_modalities":["text"]}},{"id":"a/u"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	lite, err := parseLiteLLM([]byte(`{"v":{"supports_vision":true},"t":{"supports_vision":false},"u":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		entry         Entry
		known, vision bool
	}{
		{router["a/v"], true, true}, {router["a/t"], true, false}, {router["a/u"], false, false},
		{lite["v"], true, true}, {lite["t"], true, false}, {lite["u"], false, false},
	}
	for i, c := range cases {
		if c.entry.VisionKnown != c.known || c.entry.Vision != c.vision {
			t.Errorf("case %d: %+v, want known=%v vision=%v", i, c.entry, c.known, c.vision)
		}
	}
}

func TestParseLiteLLMSkipsMalformedEntries(t *testing.T) {
	data := `{"sample_spec":{"max_input_tokens":"max tokens"},"m":{"input_cost_per_token":0.5}}`
	table, err := parseLiteLLM([]byte(data))
	if err != nil || table["m"].InputCostPerToken != 0.5 {
		t.Errorf("table = %+v, %v", table, err)
	}
}
