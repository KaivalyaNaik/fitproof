export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-[var(--bg)] flex items-center justify-center p-4">
      <div className="w-full max-w-sm animate-fade-in-up">
        <div className="flex justify-center mb-8">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/logo-full.png" alt="FitProof" style={{ height: 100, width: "auto" }} />
        </div>
        <div className="bg-[var(--surface)] border border-[var(--border)] rounded-2xl p-6 sm:p-8">
          {children}
        </div>
      </div>
    </div>
  );
}
