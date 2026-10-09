package store

import "testing"

func TestDevices(t *testing.T) {
	db := openTest(t)
	device, token, err := db.AddDevice("iPhone, Safari")
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := db.DeviceByToken(token); !ok || got.ID != device.ID {
		t.Fatalf("lookup by token: %v %v", got, ok)
	}
	if _, ok := db.DeviceByToken("wrong"); ok {
		t.Fatal("a wrong token must not match")
	}
	if err := db.RenameDevice(device.ID, "My phone"); err != nil {
		t.Fatal(err)
	}
	list, _ := db.Devices()
	if len(list) != 1 || list[0].Name != "My phone" {
		t.Fatalf("list: %v", list)
	}
	if err := db.RemoveDevice(device.ID); err != nil {
		t.Fatal(err)
	}
	if _, ok := db.DeviceByToken(token); ok {
		t.Fatal("a removed device must not match")
	}
	if err := db.RemoveDevice(device.ID); err == nil {
		t.Fatal("removing twice must be an error")
	}
}
