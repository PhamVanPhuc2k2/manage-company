"use client";

import { api, ApiError } from "@/lib/api-client";
import { useQuery } from "@tanstack/react-query";

type Ping = {
  database_time: string;
  redis_ok: boolean;
  job_queued: boolean;
  request_id: string;
  version: string;
  commit: string;
};

export default function HomePage() {
  const ping = useQuery({
    queryKey: ["ping"],
    queryFn: async () =>
      (await api.get<Ping>("/ping", { skipAuth: true })).data,
  });

  return (
    <main className="mx-auto max-w-xl px-4 py-16">
      <h1 className="text-2xl font-semibold">Manage Company - nền móng</h1>
      {ping.isPending && (
        <p className="mt-4 text-neutral-500">Đang gọi API...</p>
      )}
      {ping.error && (
        <p role="alert" className="mt-4 text-red-600">
          {ping.error instanceof ApiError
            ? ping.error.message
            : "Không kết nối được máy chủ"}
        </p>
      )}
      {ping.data && (
        <pre className="mt-4 overflow-x-auto rounded-xl bg-neutral-100 p-4 text-sm dark:bg-neutral-900">
          {JSON.stringify(ping.data, null, 2)}
        </pre>
      )}
    </main>
  );
}
