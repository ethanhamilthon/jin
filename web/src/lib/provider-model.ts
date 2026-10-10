import type { Model } from "./types";
export interface ProviderModel extends Model { provider: string; provider_name: string }
export interface Catalog { models: ProviderModel[]; errors: { provider: string; message: string }[] }
export function providerModelKey(model: ProviderModel): string {
  return JSON.stringify([model.provider, model.id]);
}
