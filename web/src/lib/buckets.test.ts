import { describe, expect, it } from "vitest";
import { bucketOf, groupRows } from "./buckets";

const now = new Date(2026, 9, 10, 15, 0).getTime();
const ago = (ms: number) => new Date(now - ms).toISOString();
const minute = 60 * 1000;
const hour = 60 * minute;

describe("bucketOf", () => {
  it("puts busy sessions in Running whatever their age", () => {
    expect(bucketOf(ago(0), true, now)).toBe("Running");
    expect(bucketOf(ago(5 * 24 * hour), true, now)).toBe("Running");
  });

  it("buckets by age", () => {
    expect(bucketOf(ago(30 * minute), false, now)).toBe("Last hour");
    expect(bucketOf(ago(2 * hour), false, now)).toBe("Last 6 hours");
    expect(bucketOf(ago(7 * hour), false, now)).toBe("Today");
    expect(bucketOf(ago(2 * 24 * hour), false, now)).toBe("This Week");
    expect(bucketOf(ago(8 * 24 * hour), false, now)).toBe("Older");
  });

  it("uses the local calendar day for Today", () => {
    const startOfToday = new Date(2026, 9, 10, 0, 0).toISOString();
    const endOfYesterday = new Date(2026, 9, 9, 23, 59).toISOString();
    expect(bucketOf(startOfToday, false, now)).toBe("Today");
    expect(bucketOf(endOfYesterday, false, now)).toBe("This Week");
  });

  it("puts sessions older than a week in Older, a week old itself too", () => {
    expect(bucketOf(ago(7 * 24 * hour - minute), false, now)).toBe("This Week");
    expect(bucketOf(ago(7 * 24 * hour), false, now)).toBe("Older");
    expect(bucketOf(ago(365 * 24 * hour), false, now)).toBe("Older");
  });

  it("treats the 6 hour and 1 hour edges as the next bucket", () => {
    expect(bucketOf(ago(hour), false, now)).toBe("Last 6 hours");
    expect(bucketOf(ago(6 * hour), false, now)).toBe("Today");
  });
});

describe("groupRows", () => {
  const row = (id: string, updated: number) => ({ id, updated_at: ago(updated) });
  const idle = () => false;

  it("returns buckets in order and drops empty ones", () => {
    const rows = [row("ancient", 30 * 24 * hour), row("old", 2 * 24 * hour), row("hour", 10 * minute), row("day", 7 * hour)];
    expect(groupRows(rows, idle, now).map((group) => group.name)).toEqual(["Last hour", "Today", "This Week", "Older"]);
  });

  it("lists sessions older than a week under Older", () => {
    const rows = [row("a", 9 * 24 * hour), row("b", 20 * 24 * hour)];
    expect(groupRows(rows, idle, now)).toEqual([{ name: "Older", rows }]);
  });

  it("places each session once, in its first matching bucket", () => {
    const rows = [row("busy", 10 * minute), row("idle", 10 * minute)];
    const groups = groupRows(rows, (r) => r.id === "busy", now);
    expect(groups).toEqual([
      { name: "Running", rows: [rows[0]] },
      { name: "Last hour", rows: [rows[1]] },
    ]);
  });

  it("keeps input order inside a bucket", () => {
    const rows = [row("a", 10 * minute), row("b", 20 * minute), row("c", 5 * minute)];
    expect(groupRows(rows, idle, now)[0].rows.map((r) => r.id)).toEqual(["a", "b", "c"]);
  });

  it("returns no groups for no rows", () => {
    expect(groupRows([], idle, now)).toEqual([]);
  });
});
