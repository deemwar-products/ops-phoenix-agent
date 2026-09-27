"use client";

const steps = [
  {
    num: "01",
    title: "Detect",
    body:
      "Queries your observability backend (Grafana, Loki, Cloud Monitoring) for errors in the last N minutes. Deduplicates by pattern and surfaces the top finding with frequency counts. Takes seconds.",
  },
  {
    num: "02",
    title: "Analyze",
    body:
      "Sends the error summary to Claude, which correlates it with recent deploys, traces, and code changes. Returns: root cause, severity, confidence score, and a fix suggestion. No more cross-referencing dashboards manually.",
  },
  {
    num: "03",
    title: "Fix",
    body:
      "Claude writes a unified diff targeting the root cause. The agent applies it to a clone of your repo, runs your tests, commits on a feature branch, and opens a PR with the fix and the full analysis.",
  },
  {
    num: "04",
    title: "Deploy + Verify",
    body:
      "Nothing deploys without your approval. Once approved, the agent triggers CI/CD, monitors the deployment, and checks post-deploy metrics. If the fix didn't work, it rolls back automatically.",
  },
];

export function HowItWorks() {
  return (
    <section id="how-it-works" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-6xl mx-auto">
        <div className="mb-16 max-w-3xl">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            How it works
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-4" style={{ color: "#1b1c19" }}>
            From anomaly to deployed fix in under 5 minutes.
          </h2>
          <p className="text-base sm:text-lg leading-relaxed" style={{ color: "#6b665e" }}>
            The same four steps an SRE would take — but automated, repeatable, and always consistent.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-x-16 gap-y-12">
          {steps.map((s) => (
            <div key={s.title}>
              <div className="flex items-baseline gap-3 mb-3">
                <span className="font-mono text-sm" style={{ color: "#735c41" }}>{s.num}</span>
                <h3 className="text-xl font-bold" style={{ color: "#1b1c19" }}>{s.title}</h3>
              </div>
              <p className="text-sm leading-relaxed pl-7" style={{ color: "#6b665e" }}>
                {s.body}
              </p>
            </div>
          ))}
        </div>

        <div
          className="mt-20 p-8 rounded-2xl"
          style={{ backgroundColor: "#f0ede6", border: "1px solid #e0d9cd" }}
        >
          <p className="text-xs font-semibold uppercase tracking-[0.2em] mb-2 font-mono" style={{ color: "#735c41" }}>
            TOTAL TIME
          </p>
          <p className="text-4xl font-bold mb-2" style={{ color: "#1b1c19" }}>&lt; 5 minutes</p>
          <p className="text-sm" style={{ color: "#6b665e" }}>
            From first anomaly signal to verified fix. Zero human intervention required if you run in autonomous mode.
            Guided mode waits for your approval before deploy.
          </p>
        </div>
      </div>
    </section>
  );
}
