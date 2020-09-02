export type Dictionary<T> = { [key: string]: T };

const SUFFIXES = {
  0: "B",
  1: "KB",
  2: "MB",
  3: "GB",
  4: "TB",
  5: "PB",
  6: "EB",
};

export function formatBytes(n: number): string {
  function formatRecursive(size: number, powerOfThousand: number): string {
    if (size > 1024) {
      return formatRecursive(size / 1024, powerOfThousand + 1);
    }
    return `${size.toFixed(2)}${SUFFIXES[powerOfThousand]}`;
  }
  return formatRecursive(n, 0);
}
