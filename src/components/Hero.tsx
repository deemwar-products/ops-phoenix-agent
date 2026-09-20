"use client";

import { TerminalPanel } from "./TerminalPanel";

export function Hero() {
  return (
    <section
      className="min-h-screen flex items-center pt-24 pb-20 overflow-hidden"
      style={{ backgroundColor: "#fdfcf8" }}
    >
      <div className="max-w-6xl mx-auto px-6 w-full">
        <div className="max-w-3xl">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-6"
            style={{ color: "#735c41" }}
          >
            Local CLI · Autonomous incident response
          </p>
          <h1
            className="text-5xl sm:text-6xl md:text-7xl leading-[1.05] tracking-tight mb-6"
            style={{ color: "#1b1c19" }}
          >
            Stop reacting to
            <br />
            <span style={{ color: "#735c41" }}>incidents. Start</span>
            <br />
            fixing them.
          </h1>
          <p
            className="text-lg sm:text-xl leading-relaxed mb-10 max-w-2xl"
            style={{ color: "#6b665e" }}
          >
            SRE Agent detects anomalies, reads your logs, finds the root cause, writes
            the code fix, opens a PR, and verifies it worked — before your engineer
            finishes pouring a coffee. Deployment requires human approval.
          </p>
          <a
            href="#install"
            className="inline-flex items-center gap-2 px-8 py-3.5 font-semibold text-sm rounded-full transition-colors"
            style={{ backgroundColor: "#735c41", color: "#fff" }}
          >
            Get started
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </a>
        </div>

        <div className="mt-16 max-w-2xl">
          <TerminalPanel label="sre-agent pipeline">
{`$ sre-agent start
› Scanning Cloud Logging for anomalies...
› Anomaly detected: 5xx rate spike on /api/resumes
› Tracing to recent deploy: auth-service v2.3.1
› Root cause: token refresh regression in PR #482
› Opening PR with fix: revert token validation path
› CI passed. Awaiting human approval to deploy.
› Verified: 5xx rate back to baseline. Incident resolved in 3m 12s.`}
          </TerminalPanel>
        </div>
      </div>
    </section>
  );
}
