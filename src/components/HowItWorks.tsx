"use client";

const steps = [
  {
    step: "01",
    title: "Detect",
    description:
      "Z-score baselines on your metrics catch anomalies before they become alerts. No threshold tuning required — it learns your traffic patterns.",
  },
  {
    step: "02",
    title: "Analyze",
    description:
      "Reads Cloud Logging entries, correlates with recent deploys and traces, and identifies the likely cause. No human detective work needed.",
  },
  {
    step: "03",
    title: "Create Issue",
    description:
      "Opens a GitHub issue with the full RCA: what happened, why, and which component is responsible. Your team gets the context, not just the alarm.",
  },
  {
    step: "04",
    title: "Create PR",
    description:
      "Writes a real code fix — not a runbook suggestion. The PR includes the diff, test plan, and a summary of the root cause for review.",
  },
  {
    step: "05",
    title: "Deploy",
    description:
      "Triggers your existing CI/CD pipeline. After passing tests, deploys to production through your normal release process.",
  },
  {
    step: "06",
    title: "Verify & Rollback",
    description:
      "Monitors the key metrics post-deploy. If the fix worked, the incident closes. If it didn&apos;t, ops-phoenix rolls back automatically.",
  },
];

export function HowItWorks() {
  return (
    <section id="how-it-works" className="py-32 px-6 border-t border-surface-border">
      <div className="max-w-6xl mx-auto">
        {/* Section header */}
        <div className="text-center mb-20">
          <p className="text-sm font-semibold uppercase tracking-widest text-accent mb-4 font-mono">
            How it works
          </p>
          <h2 className="font-bold text-4xl sm:text-5xl md:text-6xl leading-tight mb-6">
            Six steps. No humans required.
          </h2>
          <p className="text-muted text-lg max-w-2xl mx-auto">
            From anomaly detected to incident resolved — here&apos;s what
            ops-phoenix does in the time it takes your engineer to pour a coffee.
          </p>
        </div>

        {/* Steps */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {steps.map((step, i) => (
            <div
              key={step.step}
              className="group relative p-8 rounded-2xl border border-surface-border bg-surface hover:border-accent/30 transition-colors"
            >
              {/* Step number */}
              <span className="font-mono text-xs text-accent/60 mb-4 block">
                Step {step.step}
              </span>
              <h3 className="text-2xl font-bold mb-3">{step.title}</h3>
              <p className="text-muted text-sm leading-relaxed">
                {step.description}
              </p>

              {/* Connector line between cards (except last in a row) */}
              {i < steps.length - 1 && (
                <div className="hidden lg:block absolute top-1/2 -right-3 w-6 h-px bg-surface-border group-hover:bg-accent/30 transition-colors" />
              )}
            </div>
          ))}
        </div>

        {/* Outcome callout */}
        <div className="mt-16 text-center p-8 rounded-2xl border border-accent/20 bg-accent/5 dark:bg-accent/10">
          <p className="font-mono text-sm text-accent mb-2">TOTAL TIME</p>
          <p className="font-bold text-4xl sm:text-5xl mb-2">&lt; 5 minutes</p>
          <p className="text-muted">
            From first anomaly signal to verified fix. Zero human intervention
            required (if you want it).
          </p>
        </div>
      </div>
    </section>
  );
}
