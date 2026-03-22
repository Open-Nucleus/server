/** Capitalize the first character of a string. */
export function capitalize(s: string): string {
  if (s.length === 0) return s;
  return s.charAt(0).toUpperCase() + s.slice(1);
}

/** Convert a string to Title Case (each word capitalized). */
export function titleCase(s: string): string {
  return s
    .split(/[\s_-]+/)
    .map((word) => capitalize(word.toLowerCase()))
    .join(' ');
}

/** Truncate a string to `max` characters, appending ellipsis if truncated. */
export function truncate(s: string, max: number): string {
  if (s.length <= max) return s;
  return s.slice(0, max) + '\u2026';
}
