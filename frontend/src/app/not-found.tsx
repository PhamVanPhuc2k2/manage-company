"use client";

import Link from "next/link";
import { useRouter } from "next/router";

export default function NotFound() {
  const router = useRouter();
  return (
    <main className="flex min-h-[100vh] items-center justify-center bg-neutral-50 px-4 dark:bg-neutral-950">
      <div className="w-full max-w-sm text-center">
        <div className="text-5xl font-semibold tracking-tight text-neutral-300 dark:text-neutral-700">
          404
        </div>
        <h1 className="mt-3 text-lg font-semibold text-neutral-900 dark:text-neutral-100">
          Không tìm thấy trang này
        </h1>
        <p className="mt-1 text-sm text-neutral-500">
          Đường dẫn có thể đã cũ, hoặc nội dung đã bị xóa.
        </p>
        <div className="mt-6 flex justify-center gap-2">
          <button
            type="button"
            onClick={() =>
              window.history.length > 1 ? router.back() : router.push("/")
            }
            className="rounded-xl px-4 py-2 text-sm font-medium text-neutral-700 ring-1 ring-neutral-200 transition hover:bg-white dark:text-neutral-200 dark:ring-neutral-700 dark:hover:bg-neutral-900"
          >
            Quay lại
          </button>
          <Link
            href="/"
            className="rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-700"
          >
            Về trang chủ
          </Link>
        </div>
      </div>
    </main>
  );
}
