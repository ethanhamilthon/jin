package provider

import "slices"

func chatMessages(messages []Message) []Message {
	out := slices.Clone(messages)
	for i := range out {
		out[i].Native = nil
	}
	return out
}

func nativeHistory(messages []Message, kind, model string) []Message {
	out := slices.Clone(messages)
	for i := range out {
		native := out[i].Native
		if native != nil && (native.Kind != kind || native.Model != model) {
			out[i].Native = nil
		}
	}
	return out
}
