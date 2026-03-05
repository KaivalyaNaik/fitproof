export function formatPoints(value: string): string {
  const num = parseFloat(value);
  if (isNaN(num)) return value;
  return (
    new Intl.NumberFormat("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(num) + " pts"
  );
}

export function formatFines(value: string): string {
  const num = parseFloat(value);
  if (isNaN(num) || num === 0) return "0.00 pts";
  return (
    "-" +
    new Intl.NumberFormat("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(num) +
    " pts"
  );
}

export function formatDate(dateStr: string): string {
  // Parse YYYY-MM-DD without timezone conversion
  const [year, month, day] = dateStr.split("-").map(Number);
  const d = new Date(year, month - 1, day);
  return d.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function formatTimestamp(isoStr: string): string {
  return new Date(isoStr).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function statusLabel(status: string): string {
  const map: Record<string, string> = {
    active: "Active",
    completed: "Completed",
    cancelled: "Cancelled",
    draft: "Draft",
  };
  return map[status] ?? status;
}
