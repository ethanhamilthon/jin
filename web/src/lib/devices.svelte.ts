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
