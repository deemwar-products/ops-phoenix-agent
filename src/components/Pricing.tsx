"use client";

const plans = [
  {
    name: "Open",
    price: "Free",
    period: "forever",
    description: "For solo engineers, side projects, and evaluation.",
    features: [
      "Unlimited services",
      "Detect + analyze + fix",
      "All integrations (Grafana, Loki, GitHub)",
      "Agent skills (Claude Code, Codex, Cursor)",
      "MIT licensed — no restrictions",
      "Community support (GitHub Issues)",
    ],
    cta: "Install free",
    href: "#install",
    highlight: false,
  },
  {
    name: "Support",
    price: "$500",
    period: "/month",
    description: "For teams running SRE Agent in production.",
    features: [
      "Everything in Open",
      "Email support · 48h SLA",
      "Setup assistance",
      "Custom integration help",
      "Priority bug fixes",
    ],
    cta: "Get support",
    href: "mailto:admin@deemwar.com?subject=SRE%20Agent%20—%20Support%20plan%20inquiry",
    highlight: true,
  },
  {
    name: "Enterprise",
    price: "Custom",
    period: "",
    description: "For orgs with custom infra, compliance, and SLA needs.",
    features: [
      "Everything in Support",
      "Dedicated support · Custom SLA",
      "Custom integrations",
      "On-prem deployment assistance",
      "Training + onboarding",
    ],
    cta: "Talk to us",
    href: "mailto:admin@deemwar.com?subject=SRE%20Agent%20—%20Enterprise%20inquiry",
    highlight: false,
  },
];

const featureRows = [
  { key: "services", label: "Services" },
  { key: "detect", label: "Detect + analyze" },
  { key: "fix", label: "AI-powered fix generation" },
  { key: "skills", label: "Agent skills" },
  { key: "observability", label: "Grafana / Loki" },
  { key: "github", label: "GitHub integration" },
  { key: "history", label: "Run history + audit log" },
  { key: "selfHosted", label: "Self-hosted" },
  { key: "support", label: "Support" },
];

export function Pricing() {
  return (
    <section id="pricing" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-6xl mx-auto">
        <div className="mb-10 max-w-3xl">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            Pricing
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-4" style={{ color: "#1b1c19" }}>
            Free to use. Support when you need it.
          </h2>
          <p className="text-base" style={{ color: "#6b665e" }}>
            SRE Agent is MIT licensed — use it however you want, no restrictions.
            Paid tiers are for teams that need professional support, custom integrations,
            and guaranteed response times.
          </p>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-sm" style={{ color: "#1b1c19" }}>
            <thead>
              <tr style={{ borderBottom: "2px solid #e0d9cd" }}>
                <th className="text-left py-4 pr-8 font-medium text-xs uppercase tracking-[0.15em]" style={{ color: "#735c41", verticalAlign: "middle", width: "33%" }}>
                  What&apos;s included
                </th>
                {plans.map((plan) => (
                  <th
                    key={plan.name}
                    className="text-left py-4 px-6 font-medium text-xs uppercase tracking-[0.15em]"
                    style={{ color: "#735c41", verticalAlign: "middle", width: "22%" }}
                  >
                    {plan.name}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              <tr>
                <td className="py-3 pr-8" style={{ color: "#8b867f", verticalAlign: "middle" }}>
                  Price
                </td>
                {plans.map((plan) => (
                  <td key={plan.name} className="py-3 px-6" style={{ verticalAlign: "middle" }}>
                    <span className="text-2xl font-bold" style={{ color: "#1b1c19" }}>{plan.price}</span>
                    {plan.period && (
                      <span className="text-xs ml-1" style={{ color: "#6b665e" }}>{plan.period}</span>
                    )}
                  </td>
                ))}
              </tr>
              <tr>
                <td className="py-3 pr-8" style={{ color: "#8b867f", verticalAlign: "middle" }}>
                  Description
                </td>
                {plans.map((plan) => (
                  <td key={plan.name} className="py-3 px-6" style={{ color: "#6b665e", verticalAlign: "middle" }}>
                    {plan.description}
                  </td>
                ))}
              </tr>
              {featureRows.map((row, i) => (
                <tr key={row.key} style={{ borderBottom: i < featureRows.length - 1 ? "1px solid #f0ede6" : "none" }}>
                  <td className="py-3.5 pr-8" style={{ color: "#6b665e", verticalAlign: "middle" }}>
                    {row.label}
                  </td>
                  {plans.map((plan) => {
                    const val = plan.features[row.key];
                    return (
                      <td key={plan.name} className="py-3.5 px-6" style={{ verticalAlign: "middle" }}>
                        {typeof val === "boolean" ? (
                          val ? (
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" className="inline-block" style={{ color: "#735c41" }}>
                              <path d="M5 13l4 4L19 7" />
                            </svg>
                          ) : (
                            <span style={{ color: "#d5cfc4" }}>—</span>
                          )
                        ) : (
                          <span className="text-sm" style={{ color: "#1b1c19" }}>{val}</span>
                        )}
                      </td>
                    );
                  })}
                </tr>
              ))}
              <tr>
                <td className="py-5 pr-8" style={{ verticalAlign: "middle" }} />
                {plans.map((plan) => (
                  <td key={plan.name} className="py-5 px-6" style={{ verticalAlign: "middle" }}>
                    <a
                      href={plan.href}
                      className="inline-block px-6 py-2.5 rounded-full text-sm font-semibold transition-colors"
                      style={{
                        backgroundColor: plan.highlight ? "#735c41" : "transparent",
                        color: plan.highlight ? "#fff" : "#735c41",
                        border: plan.highlight ? "none" : "1px solid #d5cfc4",
                      }}
                    >
                      {plan.cta}
                    </a>
                  </td>
                ))}
              </tr>
            </tbody>
          </table>
        </div>

        <p className="mt-8 text-sm" style={{ color: "#8b867f" }}>
          Open plan is free forever — install via brew, go, npm, or pip. The source is
          MIT licensed on GitHub. Support and Enterprise plans provide email support,
          setup assistance, custom integrations, and SLAs.
          Questions:{" "}
          <a href="mailto:admin@deemwar.com" style={{ color: "#735c41", textDecoration: "underline" }}>
            admin@deemwar.com
          </a>
        </p>
      </div>
    </section>
  );
}
