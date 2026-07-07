import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Customer Portal",
  description: "Access your billing history and subscription details.",
};

export default function AccessLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}
