"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { MailCheck } from "lucide-react";

const accessSchema = z.object({
  email: z.string().email("Please enter a valid email address."),
});

type AccessFormValues = z.infer<typeof accessSchema>;

export default function PortalAccessPage() {
  const [submitted, setSubmitted] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<AccessFormValues>({
    resolver: zodResolver(accessSchema),
  });

  const onSubmit = async (data: AccessFormValues) => {
    // Mock API call
    await new Promise((resolve) => setTimeout(resolve, 800));
    setSubmitted(true);
  };

  if (submitted) {
    return (
      <div className="bg-white rounded-2xl shadow-sm border border-slate-100 p-8 text-center space-y-6">
        <div className="w-16 h-16 bg-success-600/10 text-success-600 rounded-full flex items-center justify-center mx-auto">
          <MailCheck className="w-8 h-8" />
        </div>
        <div className="space-y-2">
          <h2 className="text-2xl font-bold text-slate-900">Check your inbox</h2>
          <p className="text-slate-500 text-sm">
            We've sent a secure, passwordless link to your email. Click it to manage your subscription.
          </p>
        </div>
        <Button 
          variant="outline" 
          className="w-full"
          onClick={() => setSubmitted(false)}
        >
          Try another email
        </Button>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-slate-100 p-6 sm:p-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-slate-900 tracking-tight">
          Manage your subscription
        </h1>
        <p className="text-slate-500 text-sm mt-2">
          Enter the email address you used to subscribe to Acme Corp.
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <div className="space-y-2">
          <label className="text-sm font-medium text-slate-900" htmlFor="email">
            Email address
          </label>
          <Input
            id="email"
            type="email"
            placeholder="you@example.com"
            {...register("email")}
            className="h-12" // larger tap target for mobile
          />
          {errors.email && (
            <p className="text-sm text-danger-600">{errors.email.message}</p>
          )}
        </div>

        <Button type="submit" className="w-full h-12 text-base" disabled={isSubmitting}>
          {isSubmitting ? "Sending link..." : "Send access link"}
        </Button>
      </form>
    </div>
  );
}
