"use client";

export function Hero() {
  return (
    <section className="relative min-h-screen flex items-center justify-center pt-20 pb-32 overflow-hidden">
      <div
        className="absolute inset-0 opacity-[0.03] dark:opacity-[0.05]"
        style={{
          backgroundImage:
            "linear-gradient(var(--foreground) 1px, transparent 1px), linear-gradient(90deg, var(--foreground) 1px, transparent 1px)",
          backgroundSize: "60px 60px",
        }}
      />

      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-accent/10 dark:bg-accent/5 rounded-full blur-3xl pointer-events-none" />

      <div className="relative z-10 max-w-5xl mx-auto px-6 text-center">
        <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full border border-surface-border bg-surface text-xs font-medium text-muted mb-8">
          <span className="w-1.5 h-1.5 rounded-full bg-accent animate-pulse" />
          Now in private beta
        </div>

        <h1 className="text-5xl sm:text-6xl md:text-7xl lg:text-8xl leading-[0.95] tracking-tight mb-8">
          <span className="font-bold">Most tools tell you</span>
          <br />
          <span className="text-accent">what broke.</span>
          <br />
          <span className="font-bold">SRE Agent</span>
          <span className="text-accent"> fixes it.</span>
        </h1>

        <p className="text-lg sm:text-xl text-muted max-w-3xl mx-auto mb-10 leading-relaxed">
          An AI agent that detects incidents, reads your logs, finds the root
          cause, writes the code fix, opens a PR, deploys it, and verifies it
          worked — all while your engineer is still reading the alert.
        </p>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
          <a
            href="#"
            className="group relative inline-flex items-center gap-2 px-8 py-3.5 bg-accent text-white font-semibold rounded-full hover:bg-accent-soft transition-colors"
          >
            Get Started
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              className="group-hover:translate-x-0.5 transition-transform"
            >
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </a>
          <a
            href="https://github.com/deemwar-products/ops-phoenix-agent"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 px-8 py-3.5 border border-surface-border rounded-full hover:border-foreground/30 transition-colors font-medium"
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="currentColor"
            >
              <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
            </svg>
            View on GitHub
          </a>
        </div>

        <div className="mt-16 max-w-2xl mx-auto">
          <div className="rounded-xl border border-surface-border bg-surface overflow-hidden text-left shadow-2xl">
            <div className="flex items-center gap-2 px-4 py-3 border-b border-surface-border">
              <span className="w-3 h-3 rounded-full bg-red-400/80" />
              <span className="w-3 h-3 rounded-full bg-yellow-400/80" />
              <span className="w-3 h-3 rounded-full bg-green-400/80" />
              <span className="ml-3 text-xs text-muted font-mono">
                sre-agent pipeline
              </span>
            </div>
            <pre className="p-5 text-sm font-mono leading-relaxed overflow-x-auto text-code-text">
              <code>
                <span className="text-accent">$</span> sre-agent start{`\n`}
                <span className="text-muted">{"›"} Scanning Cloud Logging for anomalies...</span>{`\n`}
                <span className="text-muted">{"›"} Anomaly detected: 5xx rate spike on /api/resumes</span>{`\n`}
                <span className="text-muted">{"›"} Tracing to recent deploy: auth-service v2.3.1</span>{`\n`}
                <span className="text-muted">{"›"} Root cause: token refresh regression in PR #482</span>{`\n`}
                <span className="text-green-400">
                  {"›"} Opening PR with fix: revert token validation path
                </span>
                {`\n`}
                <span className="text-muted">{"›"} CI passed. Deploying to production...</span>{`\n`}
                <span className="text-green-400">
                  {"›"} Verified: 5xx rate back to baseline. Incident resolved in 3m 12s.
                </span>
              </code>
            </pre>
          </div>
        </div>
      </div>
    </section>
  );
}
