import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { SubbaLogo } from "@/components/brand/subba-logo";

export default function NotFound() {
  return (
    <div className="min-h-dvh flex flex-col items-center justify-center bg-slate-50 text-slate-900 px-6">
      <SubbaLogo size="md" showText={false} className="mb-8" />
      <h1 className="text-6xl font-extrabold tracking-tight text-slate-900 mb-4">404</h1>
      <h2 className="text-2xl font-bold tracking-tight text-slate-800 mb-2">
        Page not found
      </h2>
      <p className="text-slate-500 text-center max-w-md mb-8">
        We couldn't find the page you were looking for. It might have been moved, deleted, or never existed.
      </p>
      <Button asChild size="lg" className="bg-brand-600 hover:bg-brand-700">
        <Link href="/">
          <ArrowLeft className="h-4 w-4 mr-2" aria-hidden />
          Back to home
        </Link>
      </Button>
    </div>
  );
}
