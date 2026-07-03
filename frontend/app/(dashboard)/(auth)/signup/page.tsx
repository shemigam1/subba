"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const signupSchema = z.object({
  name: z.string().min(2, "Company name must be at least 2 characters."),
  email: z.string().email("Please enter a valid email address."),
  password: z.string().min(8, "Password must be at least 8 characters."),
});

type SignupFormValues = z.infer<typeof signupSchema>;

export default function SignupPage() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SignupFormValues>({
    resolver: zodResolver(signupSchema),
  });

  const onSubmit = async (data: SignupFormValues) => {
    setError(null);
    try {
      // Temporary mock navigation until API is wired
      router.push("/overview");
    } catch (err: any) {
      setError(err.message || "Something went wrong.");
    }
  };

  return (
    <div className="min-h-screen flex flex-row-reverse">
      {/* Left panel (visual right) - Branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-slate-900 flex-col justify-center p-12 text-white">
        <div className="max-w-md mx-auto">
          <div className="flex items-center gap-2 mb-8">
            <div className="w-8 h-8 bg-brand-600 rounded-sm" />
            <span className="font-bold text-2xl tracking-tight">Subba</span>
          </div>
          <h1 className="text-4xl font-bold mb-6">
            Get started with Subba.
          </h1>
          <p className="text-slate-400 text-lg">
            Set up your tenant account and start managing subscriptions across Africa with ease.
          </p>
        </div>
      </div>

      {/* Right panel (visual left) - Auth Form */}
      <div className="w-full lg:w-1/2 flex items-center justify-center p-8 bg-white">
        <div className="max-w-sm w-full space-y-8">
          <div>
            <h2 className="text-2xl font-bold text-slate-900">Create your account</h2>
            <p className="text-sm text-slate-500 mt-2">
              Start building your recurring revenue business
            </p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 mt-8">
            {error && (
              <div className="p-3 bg-danger-600/10 text-danger-600 rounded-md text-sm">
                {error}
              </div>
            )}

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-900" htmlFor="name">
                Company name
              </label>
              <Input
                id="name"
                type="text"
                placeholder="Acme Corp"
                {...register("name")}
              />
              {errors.name && (
                <p className="text-sm text-danger-600">{errors.name.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-900" htmlFor="email">
                Email address
              </label>
              <Input
                id="email"
                type="email"
                placeholder="jane@company.com"
                {...register("email")}
              />
              {errors.email && (
                <p className="text-sm text-danger-600">{errors.email.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-900" htmlFor="password">
                Password
              </label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                {...register("password")}
              />
              {errors.password && (
                <p className="text-sm text-danger-600">{errors.password.message}</p>
              )}
            </div>

            <Button type="submit" className="w-full" disabled={isSubmitting}>
              {isSubmitting ? "Creating account..." : "Create account"}
            </Button>
          </form>

          <p className="text-center text-sm text-slate-500">
            Already have an account?{" "}
            <Link href="/login" className="font-medium text-brand-600 hover:underline">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
