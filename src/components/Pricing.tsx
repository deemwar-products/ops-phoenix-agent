"use client";

import { ContactModal } from "./ContactModal";

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
    price: "$100",
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

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 items-start">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className="rounded-2xl overflow-hidden"
              style={{
                border: plan.highlight ? "2px solid #735c41" : "1px solid #e0d9cd",
                backgroundColor: "#ffffff",
              }}
            >
              <div className="px-6 pt-6 pb-4" style={{ borderBottom: "1px solid #f0ede6" }}>
                <p className="text-xs font-semibold uppercase tracking-[0.15em] mb-3 font-mono" style={{ color: "#735c41" }}>
                  {plan.name}
                </p>
                <div className="flex items-baseline gap-1 mb-2">
                  <span className="text-3xl font-bold" style={{ color: "#1b1c19" }}>{plan.price}</span>
                  {plan.period && (
                    <span className="text-xs" style={{ color: "#6b665e" }}>{plan.period}</span>
                  )}
                </div>
                <p className="text-xs" style={{ color: "#6b665e" }}>{plan.description}</p>
              </div>

              <div className="px-6 py-5">
                <ul className="space-y-3">
                  {plan.features.map((feature) => (
                    <li key={feature} className="flex items-start gap-2.5 text-sm" style={{ color: "#1b1c19" }}>
                      <svg
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2.5"
                        className="mt-0.5 flex-shrink-0"
                        style={{ color: "#735c41" }}
                      >
                        <path d="M5 13l4 4L19 7" />
                      </svg>
                      <span>{feature}</span>
                    </li>
                  ))}
                </ul>
              </div>

              <div className="px-6 pb-6">
                {plan.name !== "Open" ? (
                  <ContactModal
                    trigger={
                      <span className="block w-full text-center py-3 rounded-full text-sm font-semibold cursor-pointer" style={{
                        backgroundColor: plan.highlight ? "#735c41" : "transparent",
                        color: plan.highlight ? "#fff" : "#735c41",
                        border: plan.highlight ? "none" : "1px solid #d5cfc4",
                      }}>
                        {plan.cta}
                      </span>
                    }
                  />
                ) : (
                  <a
                    href={plan.href}
                    className="block w-full text-center py-3 rounded-full text-sm font-semibold transition-colors"
                    style={{
                      backgroundColor: "#735c41",
                      color: "#fff",
                    }}
                  >
                    {plan.cta}
                  </a>
                )}
              </div>
            </div>
          ))}
        </div>

        <p className="mt-10 text-sm" style={{ color: "#8b867f" }}>
          Open plan is free forever — install via Go or from source. The source is
          MIT licensed on GitHub. Support and Enterprise plans provide email support,
          setup assistance, custom integrations, and SLAs.
          Questions:{" "}
          <ContactModal
            trigger={
              <span className="cursor-pointer" style={{ color: "#735c41", textDecoration: "underline" }}>
                admin@deemwar.com
              </span>
            }
          />
        </p>
      </div>
    </section>
  );
}
