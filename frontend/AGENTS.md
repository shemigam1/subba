<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.

<!-- END:nextjs-agent-rules -->

## Change Logging Policy

- **Timeline log**: Whenever you make changes, YOU MUST ALWAYS append a brief, timestamped summary of your changes to `timeline-changes.md`.
- **Format**: Keep them short, sequential, and permanent. Do NOT overwrite other lines in the file.

## 🚢 Commit & Push Strategy

When the user says **"commit"**:

1. **Build first**: Always run the project build (`npm run build` or equivalent) before committing. If it fails, surface the error and stop — do not commit broken code.
2. **Audit uncommitted files**: Run `git status` and review everything untracked/modified.
3. **Respect gitignore**: Never force-add files that are gitignored. In particular, do NOT commit:
   - `*.m.md` files (e.g. `chat.m.md`)
   - Anything under `dev-utils/` unless the user explicitly asks for it
   - Any file matching an existing `.gitignore` rule
4. **Group commits atomically and sensibly**: Don't dump everything into one commit. Split by logical concern — one commit per feature/flow/fix. Examples:
   - DB migrations → their own commit
   - Admin-side feature → separate commit from agent-side
   - Refactor/cleanup → separate from feature work
   - Style-only token cleanup → separate from behavior changes
5. **Commit message style**: Follow the existing repo convention (scope prefixes like `feat(admin):`, `fix:`, `chore:`, `style(ui):`, `refactor:`). Short first line, no trailing period. Body only when the "why" is non-obvious.
6. **Push to main last**: After all atomic commits land locally, push to `main` in a single `git push`.
7. **Never** `--no-verify`, `--force`, `--amend` published commits, or skip hooks. If a pre-commit hook fails, fix the underlying issue and make a new commit — never bypass.

## 📋 Todo Tracking Policy

When the user mentions deferred work — phrases like "we can adjust later", "don't forget this later", "take note", "not now but eventually", "park this for now" — **you MUST immediately append the item to `todo.m.md`**.

### Format

```markdown
### YYYY-MM-DD — [Source]

- [ ] **Concise title** — One-line explanation of what needs to be done and why it was deferred.
```

### Rules

1. **Source**: Where the item came from — e.g. `Contractor`, `Self`, `Bug`, `Code Review`.
2. **One item per bullet**. No nested sub-tasks. Keep it scannable.
3. **Never delete completed items** — mark them `[x]` with a completion date comment.
4. **Append only** — never rewrite the file. New items go at the bottom.
5. The file is gitignored (`*.m.md`). It is a local working document, not committed.

## 💰 Fintech Engineering Guardrails

> [!IMPORTANT]
> This is a financial application handling other people's money. These are hard rules, not suggestions. When in doubt, choose the safer, more auditable option and ask.

### Money representation

1. **Minor units only**: Store all monetary amounts as integers in the smallest currency unit (**kobo** for NGN). Never store money as a decimal/float/double. Postgres type is `BIGINT`.
2. **No floating-point math on money. Ever.** No `float`, no JS `Number` division that can produce fractions of a unit. Compute in integer minor units.
3. **Always pair an amount with its currency.** No "naked" amounts. Default and only money-flow currency for MVP is `NGN`.
4. **Round explicitly and consistently** at well-defined boundaries (e.g. fee calculation), never implicitly.

### Ledger & integrity

5. **Financial records are append-only.** Never `UPDATE` or `DELETE` a contribution, payout, dues payment, or expense to "correct" it. Issue a reversing/adjusting entry. Status transitions (`pending → confirmed → reversed`) are allowed; rewriting history is not.
6. **Idempotency on every money mutation.** Use a client-supplied idempotency key (the prototype already generates client-side IDs — use them) so retries never double-charge or double-credit.
7. **Atomic transactions** for any operation that touches more than one row's balance/state. All-or-nothing.
8. **The server reconciles, never the client.** A contribution becomes `confirmed` only after a verified Paystack **webhook** — never on the strength of a client-side success callback. Treat all client-reported state as a hint, not truth.

### Security & secrets

9. **Never store raw card data (PCI-DSS).** Only Paystack tokens/references. No PAN, CVV, or full card number touches our DB or logs.
10. **Secret keys are server-only.** Paystack secret keys, service-role keys, and webhook secrets never reach the client bundle, never get logged.
11. **Verify all inbound webhooks** by signature before acting on them.
12. **Authorization lives at the data layer.** RLS is the real boundary; UI role-gating is cosmetic. Default-deny, scope every read/write to membership and `auth.uid()`.
13. **Least privilege & sensitive-data minimisation.** Expose PII only to those who need it (e.g. the real name behind an anonymous contribution is admin-only). Collect the minimum.

### Authorization & controls

14. **High-value actions need extra control.** Payouts above an agreed threshold require multi-party (2-of-n) admin approval.
15. **Validate server-side, always.** Client validation is UX; server/DB validation is the safeguard. Re-check amounts, ownership, and limits on the server.
16. **Amounts must be positive.** `CHECK (amount_kobo > 0)` on every money column; reject zero/negative at the edge too.

### Audit & time

17. **Immutable audit trail.** Every money movement and admin action records who, what, and when, in a way that cannot be silently altered.
18. **UTC + `timestamptz` everywhere.** Store timestamps with timezone; convert display strings to ISO on ingest. Never store money-event times as naive local strings.
19. **No destructive operations without a backup and explicit go-ahead.** No `DROP`/`TRUNCATE`/mass `UPDATE` on production financial tables without confirmation.

## ??? Security & Vulnerability Auditing

1. **Mandatory Security Audit**: Before concluding any feature implementation or refactoring, you MUST perform a proactive security vulnerability check on the changes you just made.
2. **Impact Analysis**: Evaluate how your new code interacts with existing systems (e.g., RLS bypasses, missing authorization checks on related data, type coercion flaws, missing server-side input validation).
3. **Fix Forward**: If you identify a vulnerability during your self-audit, immediately document it and propose/implement the patches before marking the task as complete. Never leave known exploits in the codebase.
