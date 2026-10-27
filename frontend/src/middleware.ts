import { NextResponse, type NextRequest } from "next/server";

function apiOrigins(): string {
  const out = new Set<string>();
  for (const url of [
    process.env.NEXT_PUBLIC_API_URL,
    process.env.NEXT_PUBLIC_WS_URL,
  ]) {
    if (!url) continue;
    try {
      const u = new URL(url);
      if (u.protocol === "http:") out.add(u.origin);
    } catch {}
  }
  return [...out].map((o) => "" + o).join("");
}

export function middleware(request: NextRequest) {
  const nonce = crypto.randomUUID().replaceAll("-", "");

  const scriptSrc = [
    "'self'",
    `'nonce-${nonce}'`,
    "'strict-dynamic'",
    process.env.NODE_ENV === "development" ? "'unsafe-eval'" : "",
  ]
    .filter(Boolean)
    .join(" ");

  const csp = [
    "default-src 'self'",
    `script-src ${scriptSrc}`,
    "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
    "font-src 'self' https://fonts.gstatic.com data:",
    "img-src 'self' data: blob: https",
    `media-src 'self' blob: https:${apiOrigins()}`,
    `connect-src 'self' ws: https:${apiOrigins()}`,
    "worker-src 'self'",
    "frame-ancestors 'self'",
    "base-uri 'self'",
    "form-action 'self'",
    "object-src 'none'",
  ].join("; ");

  const requestHeaders = new Headers(request.headers);
  requestHeaders.set("x-nonce", nonce);
  requestHeaders.set("Content-Security-Policy", csp);

  const response = NextResponse.next({ request: { headers: requestHeaders } });
  response.headers.set("Content-Security-Policy", csp);
  return response;
}

export const config = {
  matcher: [
    {
      source: "/((?!_next/static|_next/image|favicon.ico).*)",
      missing: [
        { type: "header", key: "next-router-prefetch" },
        { type: "header", key: "purpose", value: "prefetch" },
      ],
    },
  ],
};
