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
            className="text-[11px] font-semibold text-zinc-400 uppercase tracking-widest"
          >
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={id}
          className={[
            "block w-full rounded-xl border px-3.5 py-2.5 text-sm",
            "text-zinc-900 placeholder-zinc-400 bg-white",
            "transition-colors duration-150",
            error
              ? "border-red-400 focus:border-red-500"
              : "border-zinc-200 focus:border-zinc-900",
            "focus:outline-none focus:ring-0",
            "disabled:bg-zinc-50 disabled:text-zinc-400 disabled:cursor-not-allowed",
            className,
          ].join(" ")}
          {...props}
        />
        {error && <p className="text-xs text-red-500">{error}</p>}
      </div>
    );
  }
);
Input.displayName = "Input";
