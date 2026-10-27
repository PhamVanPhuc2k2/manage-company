"use client";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="vi">
      <body className="bg-white text-neutral-900">
        <main className="flex min-h-screen items-center justify-center px-4">
          <div className="max-w-md text-center">
            <h1 className="text-xl font-semibold">Đã có lỗi xảy ra</h1>
            <p className="mt-2 text-sm text-neutral-500">
              Hệ thống gặp sự cố ngoài dự kiến. Vui lòng thử lại.
            </p>

            {error.digest && (
              <p className="mt-2 font-mono text-xs text-neutral-400">
                Mã lỗi: {error.digest}
              </p>
            )}
            <button className="mt-6 rounded bg-neutral-900 px-4 py-2 text-sm font-medium text-white">
              Thử lại
            </button>
          </div>
        </main>
      </body>
    </html>
  );
}
