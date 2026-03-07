import Image from "next/image";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-zinc-50 flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="flex justify-center mb-8">
          <Image src="/logo-full.png" alt="FitProof" width={140} height={38} priority />
        </div>
        <div className="bg-white rounded-2xl shadow-sm ring-1 ring-zinc-950/[0.06] p-6 sm:p-8">
          {children}
        </div>
      </div>
    </div>
  );
}
