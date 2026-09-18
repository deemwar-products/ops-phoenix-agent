"use client";

const features = [
  {
    title: "Anomaly Detection",
    description:
      "Z-score baselines on all your metrics — no threshold tuning. Learns your traffic patterns and flags deviations before they trigger cascading failures.",
    tag: "proactive",
  },
  {
    title: "Root Cause Analysis",
    description:
      "Reads Cloud Logging entries, correlates with recent deploys and traces, and explains what happened in plain English. Plus a confidence score.",
    tag: "intelligent",
  },
  {
    title: "Code-Level Fixes",
    description:
      "Writes actual code diffs — not runbook suggestions. Understands your codebase, proposes a minimal fix, and explains the reasoning.",
    tag: "autonomous",
  },
  {
    title: "GitHub Integration",
    description:
      "Opens a PR with the fix, links the RCA, and triggers your CI/CD pipeline. Auto-merges on passing tests — or waits for your approval.",
    tag: "native",
  },
  {
    title: "Outcome Verification",
    description:
      "Monitors key metrics post-deploy. If the fix worked, closes the incident. If not, rolls back automatically. Close the loop, not just the ticket.",
    tag: "verified",
  },
  {
    title: "Guided or Autonomous",
    description:
      "You choose the autonomy level. Approve every step, approve fixes only, or let it run end-to-end. The pipeline is the same — your hand is optional.",
    tag: "flexible",
  },
];

export function Features() {
  return (
    <section id="features" className="py-32 px-6 border-t border-surface-border">
      <div className="max-w-6xl mx-auto">
        {/* Section header */}
        <div className="text-center mb-20">
          <p className="text-sm font-semibold uppercase tracking-widest text-accent mb-4 font-mono">
            What you get
          </p>
          <h2 className="font-display text-4xl sm:text-5xl md:text-6xl leading-tight mb-6">
            Six capabilities.
            <br />
            <span className="text-accent">Zero alert fatigue.</span>
          </h2>
          <p className="text-muted text-lg max-w-2xl mx-auto">
            Every piece you need to go from incident detected to incident
            resolved — without waking up your engineer.
          </p>
        </div>

        {/* Feature grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {features.map((feature) => (
            <div
              key={feature.title}
              className="p-8 rounded-2xl border border-surface-border bg-surface hover:border-accent/30 transition-colors group"
            >
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-xl font-semibold">{feature.title}</h3>
                <span className="text-[10px] font-mono uppercase tracking-widest text-muted bg-background px-2 py-1 rounded-full">
                  {feature.tag}
                </span>
              </div>
              <p className="text-muted text-sm leading-relaxed">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
