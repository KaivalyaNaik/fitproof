type BadgeVariant =
  | "active"
  | "completed"
  | "cancelled"
  | "draft"
  | "host"
  | "cohost"
  | "participant";

const variantMap: Record<BadgeVariant, string> = {
  active: "bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200/60",
  completed: "bg-sky-50 text-sky-700 ring-1 ring-sky-200/60",
  cancelled: "bg-red-50 text-red-600 ring-1 ring-red-200/60",
  draft: "bg-amber-50 text-amber-700 ring-1 ring-amber-200/60",
  host: "bg-violet-50 text-violet-700 ring-1 ring-violet-200/60",
  cohost: "bg-purple-50 text-purple-700 ring-1 ring-purple-200/60",
  participant: "bg-zinc-100 text-zinc-600 ring-1 ring-zinc-200/60",
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
      className={`inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-medium tracking-wide ${variantMap[variant]}`}
    >
      {label}
    </span>
  );
}
