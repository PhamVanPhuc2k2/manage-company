import { QueryProvider } from "@/lib/query-provider";
import { THEME_SCRIPT } from "@/lib/theme";
import { Metadata } from "next";
import { headers } from "next/headers";

export const metadata: Metadata = {
  title: "Quản lý công ty",
  description: "Hệ thống quản lý nhân sự, dự án, chấm công và lương",
  icons: { apple: "/icons/logo.png" },
  appleWebApp: { capable: true, title: "TNP", statusBarStyle: "default" },
};

export default async function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const nonce = (await headers()).get("x-nonce") ?? undefined;

  return (
    <html lang="vi" suppressHydrationWarning>
      <head>
        <script
          nonce={nonce}
          dangerouslySetInnerHTML={{ __html: THEME_SCRIPT }}
        />
      </head>
      <body className="bg-white text-neutral-900 antialiased dark:bg-neutral-950 dark:text-neutral-100">
        <QueryProvider>{children}</QueryProvider>
      </body>
    </html>
  );
}
