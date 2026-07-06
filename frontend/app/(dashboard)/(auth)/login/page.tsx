"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import { setTenantToken } from "@/lib/auth/token";

const loginSchema = z.object({
  email: z.string().email("Please enter a valid email address."),
  password: z.string().min(8, "Password must be at least 8 characters."),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginFormValues) => {
    setError(null);
    const { data: res, error } = await api.POST("/auth/login", { body: data });
    if (error) {
      setError((error as { message?: string })?.message ?? "Invalid email or password.");
      return;
    }
    const token = (res as { token?: string })?.token;
    if (token) setTenantToken(token);
    router.push("/overview");
  };

  return (
    <div className="min-h-screen flex">
      {/* Left panel - Branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-brand-600 flex-col justify-center p-12 text-white">
        <div className="max-w-md mx-auto">
          <div className="flex items-center gap-2 mb-8">
            <div className="w-8 h-8 bg-white rounded-sm" />
            <span className="font-bold text-2xl tracking-tight">Subba</span>
          </div>
          <h1 className="text-4xl font-bold mb-6">
            Smarter subscriptions for African businesses.
          </h1>
          <p className="text-brand-50 text-lg">
            Manage your recurring revenue, reduce churn, and offer cardless auto-renewals built natively on Nomba's rails.
          </p>
        </div>
      </div>

      {/* Right panel - Auth Form */}
      <div className="w-full lg:w-1/2 flex items-center justify-center p-8 bg-slate-50">
        <div className="max-w-sm w-full space-y-8">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-slate-900">Welcome back</h2>
            <p className="text-sm text-slate-500 mt-2">
              Sign in to manage your Subba account
            </p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 mt-8">
            {error && (
              <div className="p-3 bg-danger-600/10 text-danger-600 rounded-md text-sm">
                {error}
              </div>
            )}
            
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
              <div className="flex items-center justify-between">
                <label className="text-sm font-medium text-slate-900" htmlFor="password">
                  Password
                </label>
                <Link href="#" className="text-sm font-medium text-brand-600 hover:underline">
                  Forgot password?
                </Link>
              </div>
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
              {isSubmitting ? "Signing in..." : "Sign in"}
            </Button>
          </form>

          <p className="text-center text-sm text-slate-500">
            Don't have an account?{" "}
            <Link href="/signup" className="font-medium text-brand-600 hover:underline">
              Sign up
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
