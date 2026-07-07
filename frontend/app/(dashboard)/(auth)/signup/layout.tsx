import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Sign up",
  description: "Create your Subba account to start managing subscriptions natively on Nomba.",
};

export default function SignupLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}
