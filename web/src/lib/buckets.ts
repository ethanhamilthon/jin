export const bucketNames = ["Running", "Last hour", "Last 6 hours", "Today", "This Week"] as const;

export type Bucket = (typeof bucketNames)[number];

export interface Group<T> {
  name: Bucket;
  rows: T[];
}

const hour = 3600 * 1000;

export function bucketOf(updatedAt: string, busy: boolean, now: number): Bucket {
  if (busy) return "Running";
  const time = new Date(updatedAt).getTime();
  if (now - time < hour) return "Last hour";
  if (now - time < 6 * hour) return "Last 6 hours";
  const midnight = new Date(now);
  midnight.setHours(0, 0, 0, 0);
  return time >= midnight.getTime() ? "Today" : "This Week";
}

export function groupRows<T extends { updated_at: string }>(rows: T[], busy: (row: T) => boolean, now: number): Group<T>[] {
  const groups: Group<T>[] = bucketNames.map((name) => ({ name, rows: [] }));
  for (const row of rows) groups[bucketNames.indexOf(bucketOf(row.updated_at, busy(row), now))].rows.push(row);
  return groups.filter((group) => group.rows.length > 0);
}
