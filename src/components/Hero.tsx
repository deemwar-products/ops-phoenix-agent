"use client";

export function Hero() {
  return (
    <section
      className="min-h-screen flex items-center pt-24 pb-20 overflow-hidden"
      style={{ backgroundColor: "#fdfcf8" }}
    >
      <div className="max-w-6xl mx-auto px-6 w-full">
        <div className="max-w-3xl">
          <div className="flex flex-wrap gap-2 mb-8">
            {["CLI", "Agent skills", "Self-hosted", "MIT licensed"].map((tag) => (
              <span
                key={tag}
                className="text-[11px] font-mono font-medium uppercase tracking-[0.15em] px-3 py-1.5 rounded-full"
                style={{ backgroundColor: "#f0ede6", color: "#735c41", border: "1px solid #e0d9cd" }}
              >
                {tag}
              </span>
            ))}
          </div>

          <h1
            className="text-5xl sm:text-6xl md:text-7xl leading-[1.05] tracking-tight mb-6"
            style={{ color: "#1b1c19" }}
          >
            From alert to PR.
            <br />
            <span style={{ color: "#735c41" }}>No manual debugging.</span>
          </h1>

          <p
            className="text-lg sm:text-xl leading-relaxed mb-10 max-w-2xl"
            style={{ color: "#6b665e" }}
          >
            SRE Agent queries your logs, finds the root cause, writes a code fix,
            and opens a PR — while your engineer is still reading the alert.
            Runs locally. No cloud. MIT licensed.
          </p>

          <div className="flex flex-wrap gap-4">
            <a
              href="#install"
              className="inline-flex items-center gap-2 px-8 py-3.5 font-semibold text-sm rounded-full transition-colors"
              style={{ backgroundColor: "#735c41", color: "#fff" }}
            >
              Install
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <path d="M5 12h14M12 5l7 7-7 7" />
              </svg>
            </a>
            <a
              href="#how-it-works"
              className="inline-flex items-center gap-2 px-8 py-3.5 font-semibold text-sm rounded-full transition-colors"
              style={{ color: "#735c41", border: "1px solid #d5cfc4" }}
            >
              See how it works
            </a>
          </div>

          <div className="flex items-center gap-4 mt-6">
            <p className="text-xs" style={{ color: "#8b867f" }}>
              go install · from source — MIT licensed
            </p>
            <a
              href="https://github.com/deemwar-products/ops-phoenix-agent"
              target="_blank"
              rel="noopener noreferrer"
              className="text-xs font-mono hover:opacity-70 transition-opacity"
              style={{ color: "#735c41" }}
            >
              View on GitHub →
            </a>
          </div>
        </div>

        <div className="mt-16 max-w-2xl">
          <div
            className="rounded-xl overflow-hidden"
            style={{ backgroundColor: "#0f1117", border: "1px solid rgba(255,255,255,0.08)" }}
          >
            <div
              className="flex items-center gap-2 px-4 py-3"
              style={{ borderBottom: "1px solid rgba(255,255,255,0.06)" }}
            >
              <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#ff5f57" }} />
              <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#febc2e" }} />
              <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#28c840" }} />
              <span className="ml-3 text-xs font-mono" style={{ color: "rgba(255,255,255,0.25)" }}>
                terminal
              </span>
            </div>
            <pre className="p-5 text-sm font-mono leading-7 overflow-x-auto" style={{ color: "#c9d1d9" }}>
              <span style={{ color: "#735c41" }}>$</span> sre-agent full-cycle --duration 1h{'\n'}
              <span style={{ color: "#8b949e" }}>
                {'›'} Scanning Grafana Loki for errors...{'\n'}
                {'›'} Found 3 errors. P1: 5xx spike /api/resumes{'\n\n'}
              </span>
              <span style={{ color: "#735c41" }}>$</span> sre-agent analyze --finding 0{'\n'}
              <span style={{ color: "#8b949e" }}>
                {'›'} Root cause: token refresh regression in auth-service v2.3.1{'\n\n'}
              </span>
              <span style={{ color: "#735c41" }}>$</span> sre-agent fix --finding 0 --approve{'\n'}
              <span style={{ color: "#3fb950" }}>
                {'›'} PR #482 opened: revert token validation path{'\n'}
                {'›'} CI passed. Team reviews and deploys via their pipeline.{'\n'}
              </span>
            </pre>
          </div>
        </div>
      </div>
    </section>
  );
}
