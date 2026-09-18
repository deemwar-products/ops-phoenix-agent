"use client";

const plans = [
  {
    name: "Open Source",
    price: "Free",
    description: "Self-hosted. Full features. Community support.",
    features: [
      "All core features",
      "GCP Cloud Logging + Monitoring",
      "GitHub integration",
      "Guided mode (human approval)",
      "Community support",
    ],
    cta: "Star on GitHub",
    ctaHref: "https://github.com/deemwar-products/ops-phoenix-agent",
    highlighted: false,
  },
  {
    name: "Pro",
    price: "Coming Soon",
    description: "Managed deployment. Advanced analytics. Priority support.",
    features: [
      "Everything in Open Source",
      "Autonomous mode (no approvals)",
      "Multi-repo / multi-service",
      "Advanced RCA with traces",
      "Priority support",
      "Custom integrations",
    ],
    cta: "Join Waitlist",
    ctaHref: "#",
    highlighted: true,
  },
];

export function Pricing() {
  return (
    <section id="pricing" className="py-32 px-6 border-t border-surface-border">
      <div className="max-w-4xl mx-auto">
        {/* Section header */}
        <div className="text-center mb-20">
          <p className="text-sm font-semibold uppercase tracking-widest text-accent mb-4 font-mono">
            Pricing
          </p>
          <h2 className="font-display text-4xl sm:text-5xl md:text-6xl leading-tight mb-6">
            Start free.
            <br />
            <span className="text-accent">Scale when you&apos;re ready.</span>
          </h2>
        </div>

        {/* Plans */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className={`relative p-8 rounded-2xl border ${
                plan.highlighted
                  ? "border-accent/40 bg-accent/5 dark:bg-accent/10"
                  : "border-surface-border bg-surface"
              }`}
            >
              {plan.highlighted && (
                <span className="absolute -top-3 left-8 text-xs font-mono uppercase tracking-widest text-accent bg-background px-3 py-1 border border-surface-border rounded-full">
                  Recommended
                </span>
              )}
              <h3 className="text-2xl font-display mb-1">{plan.name}</h3>
              <p className="text-4xl font-semibold mb-2">{plan.price}</p>
              <p className="text-muted text-sm mb-6">{plan.description}</p>

              <ul className="space-y-3 mb-8">
                {plan.features.map((feature) => (
                  <li
                    key={feature}
                    className="flex items-start gap-2 text-sm text-muted"
                  >
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2.5"
                      className="text-accent shrink-0 mt-0.5"
                    >
                      <path d="M5 13l4 4L19 7" />
                    </svg>
                    {feature}
                  </li>
                ))}
              </ul>

              <a
                href={plan.ctaHref}
                target="_blank"
                rel="noopener noreferrer"
                className={`block text-center py-3 rounded-full font-semibold text-sm transition-colors ${
                  plan.highlighted
                    ? "bg-accent text-background hover:bg-accent-soft"
                    : "border border-surface-border hover:border-foreground/30"
                }`}
              >
                {plan.cta}
              </a>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
