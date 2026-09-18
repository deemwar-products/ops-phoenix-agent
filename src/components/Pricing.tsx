"use client";

const plans = [
  {
    name: "Starter",
    price: "$499/mo",
    priceNote: "per month, billed annually",
    description: "For teams getting started with autonomous incident response.",
    features: [
      "Up to 3 services",
      "Anomaly detection + RCA",
      "GitHub PR integration",
      "Guided mode (approval required)",
      "Email support",
    ],
    cta: "Get Started",
    ctaHref: "#",
    highlighted: false,
  },
  {
    name: "Enterprise",
    price: "Custom",
    priceNote: "tailored to your infra",
    description: "For teams running at scale. Full autonomy, full support.",
    features: [
      "Unlimited services",
      "Autonomous mode (no approvals)",
      "Multi-repo / multi-service",
      "Advanced RCA with traces",
      "Custom integrations",
      "Dedicated support + SLA",
    ],
    cta: "Contact Sales",
    ctaHref: "#",
    highlighted: true,
  },
];

export function Pricing() {
  return (
    <section id="pricing" className="py-32 px-6 border-t border-surface-border">
      <div className="max-w-4xl mx-auto">
        <div className="text-center mb-20">
          <p className="text-sm font-semibold uppercase tracking-widest text-accent mb-4 font-mono">
            Pricing
          </p>
          <h2 className="text-4xl sm:text-5xl md:text-6xl leading-tight mb-6">
            Start small.
            <br />
            <span className="text-accent">Scale when you need to.</span>
          </h2>
        </div>

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
              <h3 className="text-2xl font-bold mb-1">{plan.name}</h3>
              <p className="text-4xl font-bold mb-1">{plan.price}</p>
              <p className="text-muted text-xs mb-2">{plan.priceNote}</p>
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
                className={`block text-center py-3 rounded-full font-semibold text-sm transition-colors ${
                  plan.highlighted
                    ? "bg-accent text-white hover:bg-accent-soft"
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
