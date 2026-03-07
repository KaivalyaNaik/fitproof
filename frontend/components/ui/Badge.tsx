type BadgeVariant =
  | "active"
  | "completed"
  | "cancelled"
  | "draft"
  | "host"
  | "cohost"
  | "participant";

const variantMap: Record<BadgeVariant, string> = {
  active:
    "bg-[var(--success-dim)] text-[var(--success)] ring-1 ring-[var(--success)]/20",
  completed:
    "bg-sky-950/40 text-sky-400 ring-1 ring-sky-500/20",
  cancelled:
    "bg-[var(--danger-dim)] text-[var(--danger)] ring-1 ring-[var(--danger)]/20",
  draft:
    "bg-[var(--warning-dim)] text-[var(--warning)] ring-1 ring-[var(--warning)]/20",
  host:
    "bg-[var(--accent-dim)] text-[var(--accent)] ring-1 ring-[var(--accent)]/20",
  cohost:
    "bg-violet-950/40 text-violet-400 ring-1 ring-violet-500/20",
  participant:
    "bg-[var(--surface-raised)] text-[var(--text-muted)] ring-1 ring-[var(--border)]",
};

export function Badge({
  variant,
  label,
}: {
  variant: BadgeVariant;
  label: string;
}) {
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-semibold tracking-wide ${variantMap[variant]}`}
    >
      {label}
    </span>
  );
}
