import { InputHTMLAttributes, forwardRef } from "react";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, className = "", id, ...props }, ref) => {
    return (
      <div className="flex flex-col gap-1.5">
        {label && (
          <label
            htmlFor={id}
            className="text-[10px] font-semibold text-[var(--text-muted)] uppercase tracking-widest"
          >
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={id}
          className={[
            "block w-full rounded-xl border px-3.5 py-2.5 text-sm",
            "bg-[var(--surface)] text-[var(--text)] placeholder-[var(--text-dim)]",
            "transition-colors duration-150",
            error
              ? "border-[var(--danger)] focus:border-[var(--danger)]"
              : "border-[var(--border)] focus:border-[var(--accent)]",
            "focus:outline-none focus:ring-0",
            "disabled:opacity-40 disabled:cursor-not-allowed",
            className,
          ].join(" ")}
          {...props}
        />
        {error && <p className="text-xs text-[var(--danger)]">{error}</p>}
      </div>
    );
  }
);
Input.displayName = "Input";
