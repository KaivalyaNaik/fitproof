export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-zinc-50 flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <span className="text-zinc-900 font-semibold text-lg tracking-tight">
            FitProof
          </span>
        </div>
        <div className="bg-white rounded-2xl shadow-sm ring-1 ring-zinc-950/[0.06] p-6 sm:p-8">
          {children}
        </div>
      </div>
    </div>
  );
}
