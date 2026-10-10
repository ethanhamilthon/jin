package web

import "jin/internal/sources"

type catalogModelView struct {
	modelView
	Provider     string `json:"provider"`
	ProviderName string `json:"provider_name"`
}

func (s *server) catalogView(catalog sources.Catalog) any {
	models := []catalogModelView{}
	prices := s.m.Prices()
	for _, choice := range catalog.Models {
		model := modelView{ID: choice.ID}
		if entry, ok := prices.Lookup(choice.ID); ok {
			model.Context = entry.MaxInputTokens
			model.Input, model.Output = entry.InputCostPerToken*1e6, entry.OutputCostPerToken*1e6
			model.Reasoning, model.Vision = entry.ReasoningKnown && entry.Reasoning, entry.VisionKnown && entry.Vision
		}
		models = append(models, catalogModelView{modelView: model, Provider: choice.Provider, ProviderName: choice.ProviderName})
	}
	return struct {
		Models []catalogModelView `json:"models"`
		Errors []sources.Failure  `json:"errors"`
	}{models, catalog.Errors}
}
