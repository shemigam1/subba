import {
  AlertTriangle,
  CheckCircle2,
  CircleDashed,
  Clock,
  MinusCircle,
  XCircle,
  type LucideIcon,
} from "lucide-react"

import { cn } from "@/lib/utils"

// Status is never conveyed by color alone: every badge is color + icon + label
// (DESIGN_BRIEF "status → visual map").
type Tone = "success" | "warning" | "danger" | "info" | "slate"

const tones: Record<Tone, string> = {
  success: "bg-green-50 text-green-700 border-green-200",
  warning: "bg-amber-50 text-amber-700 border-amber-200",
  danger: "bg-red-50 text-red-700 border-red-200",
  info: "bg-blue-50 text-blue-700 border-blue-200",
  slate: "bg-slate-50 text-slate-600 border-slate-200",
}

const statusMap: Record<string, { tone: Tone; label: string; icon: LucideIcon }> = {
  // subscription
  active: { tone: "success", label: "Active", icon: CheckCircle2 },
  past_due: { tone: "warning", label: "Past due", icon: AlertTriangle },
  incomplete: { tone: "info", label: "Incomplete", icon: Clock },
  canceled: { tone: "slate", label: "Canceled", icon: MinusCircle },
  unpaid: { tone: "danger", label: "Unpaid", icon: XCircle },
  // invoice
  paid: { tone: "success", label: "Paid", icon: CheckCircle2 },
  open: { tone: "info", label: "Open", icon: Clock },
  void: { tone: "slate", label: "Void", icon: MinusCircle },
  uncollectible: { tone: "danger", label: "Uncollectible", icon: XCircle },
  draft: { tone: "slate", label: "Draft", icon: CircleDashed },
}

export function StatusBadge({ status, className }: { status?: string; className?: string }) {
  const s = (status && statusMap[status]) || {
    tone: "slate" as Tone,
    label: status ?? "Unknown",
    icon: CircleDashed,
  }
  const Icon = s.icon
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium",
        tones[s.tone],
        className
      )}
    >
      <Icon className="h-3.5 w-3.5" aria-hidden />
      {s.label}
    </span>
  )
}
