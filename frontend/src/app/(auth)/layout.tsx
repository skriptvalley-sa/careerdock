/**
 * Auth route group layout — centred card layout for login, register, etc.
 * No sidebar, no header — just the form centred on screen.
 */
export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-surface px-4">
      <div className="w-full max-w-md">{children}</div>
    </div>
  );
}
