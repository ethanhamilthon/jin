package headless

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func formatPrice(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s
}

func printModels(w io.Writer, entries []modelEntry, format string) error {
	if format == "json" {
		if entries == nil {
			entries = []modelEntry{}
		}
		return json.NewEncoder(w).Encode(entries)
	}
	for _, e := range entries {
		in, out, ctxWin := "", "", ""
		if e.InputPerMTok != nil {
			in = formatPrice(*e.InputPerMTok)
		}
		if e.OutputPerMTok != nil {
			out = formatPrice(*e.OutputPerMTok)
		}
		if e.ContextWindow != nil {
			ctxWin = strconv.Itoa(*e.ContextWindow)
		}
		if e.Provider != "" {
			fmt.Fprintf(w, "%s\t", e.Provider)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", e.ID, in, out, ctxWin, strings.Join(e.Efforts, ","))
	}
	return nil
}
