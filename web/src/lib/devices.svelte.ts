import { get } from "./api";
import type { Device } from "./types";

export const devices = $state<{ remote: boolean; list: Device[] }>({ remote: false, list: [] });

export async function loadDevices() {
  const data = await get<{ remote: boolean; devices: Device[] }>("/api/devices");
  devices.remote = data.remote;
  devices.list = data.devices;
}

// others counts the devices online right now, not this page.
export const others = () => devices.list.filter((d) => d.online && !d.current).length;

// current is this page's own device id, the same value the server puts in
// the origin of an event that only its initiator acts on.
export const current = () => devices.list.find((d) => d.current)?.id ?? "";
