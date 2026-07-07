import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { MockProvider } from "@/components/mock-provider";
import { QueryProvider } from "@/components/query-provider";

const inter = Inter({
  variable: "--font-sans",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    template: "%s | Subba",
    default: "Subba | Smarter subscriptions for African businesses",
  },
  description: "An event-driven, fault-tolerant recurring billing layer built natively on Nomba's payment primitives.",
  metadataBase: new URL(process.env.NEXT_PUBLIC_DASHBOARD_URL || "http://localhost:3000"),
  openGraph: {
    title: "Subba",
    description: "Smarter subscriptions for African businesses.",
    type: "website",
    siteName: "Subba",
  },
  twitter: {
    card: "summary_large_image",
    title: "Subba",
    description: "Smarter subscriptions for African businesses.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${inter.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col font-sans">
        <MockProvider>
          <QueryProvider>{children}</QueryProvider>
        </MockProvider>
      </body>
    </html>
  );
}
