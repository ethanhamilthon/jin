package pricing

import "testing"

func TestLookupDeterministicPriority(t *testing.T) {
	table := Table{
		"zebra/model-a":     {InputCostPerToken: 100},
		"deepseek/model-a":  {InputCostPerToken: 90},
		"google/model-a":    {InputCostPerToken: 80},
		"anthropic/model-a": {InputCostPerToken: 70},
		"openai/model-a":    {InputCostPerToken: 60},
	}

	// Repeated lookups must deterministically choose openai
	for i := 0; i < 20; i++ {
		entry, ok := table.Lookup("model-a")
		if !ok || entry.InputCostPerToken != 60 {
			t.Fatalf("run %d: Lookup(model-a) = %+v, %v; want cost 60 (openai)", i, entry, ok)
		}
	}

	delete(table, "openai/model-a")
	entry, ok := table.Lookup("model-a")
	if !ok || entry.InputCostPerToken != 70 {
		t.Fatalf("Lookup without openai = %+v, %v; want cost 70 (anthropic)", entry, ok)
	}

	delete(table, "anthropic/model-a")
	entry, ok = table.Lookup("model-a")
	if !ok || entry.InputCostPerToken != 80 {
		t.Fatalf("Lookup without anthropic = %+v, %v; want cost 80 (google)", entry, ok)
	}

	delete(table, "google/model-a")
	entry, ok = table.Lookup("model-a")
	if !ok || entry.InputCostPerToken != 90 {
		t.Fatalf("Lookup alphabetical = %+v, %v; want cost 90 (deepseek)", entry, ok)
	}
}

func TestLookupExactPrecedence(t *testing.T) {
	table := Table{
		"model-b":        {InputCostPerToken: 1},
		"openai/model-b": {InputCostPerToken: 2},
	}
	entry, ok := table.Lookup("model-b")
	if !ok || entry.InputCostPerToken != 1 {
		t.Fatalf("exact id must win: got %+v, want cost 1", entry)
	}
}
