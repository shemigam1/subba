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
import { SubbaLogo } from "@/components/brand/subba-logo";

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
      const { data: res, error } = await api.POST("/auth/signup", { body: data });
      if (error) {
        setError((error as { message?: string })?.message ?? "Could not create account.");
        return;
      }
      const token = res?.token;
      if (token) setTenantToken(token);
      router.push("/overview");
    } catch (err: any) {
      setError(err.message || "Network error. Please check your connection or CORS settings.");
    }
  };

  return (
    <div className="min-h-screen flex flex-row-reverse">
      {/* Left panel (visual right) - Branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-slate-50 flex-col justify-center items-center p-12 border-l border-slate-200">
        <div className="max-w-md mx-auto text-center">
          <SubbaLogo className="mb-12" />
          <h1 className="text-4xl font-bold text-slate-900 mb-6">
            Get started with Subba.
          </h1>
          <p className="text-slate-500 text-lg">
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
