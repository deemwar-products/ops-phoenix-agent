"use client";

import { ContactModal } from "./ContactModal";

const plans = [
  {
    name: "Open",
    price: "Free",
    period: "forever",
    description: "Bring your own keys. Run it on your machine or in your own cloud.",
    features: [
      "Unlimited services — detect, analyze, fix, PR",
      "All integrations (Grafana, Loki, GitHub)",
      "Agent skills (Claude Code, Codex, Cursor)",
      "Bring your own keys · laptop, VM, or your cloud",
      "MIT licensed · community support on GitHub",
    ],
    cta: "Install free",
    href: "#install",
    highlight: false,
  },
  {
    name: "Setup",
    price: "$300",
    period: "one-time",
    description: "We get SRE Agent running against your stack. You watch, you learn, you own it.",
    features: [
      "Install on your machine or your cloud VM",
      "Grafana / Loki + GitHub connected",
      "First working cycle: alert → PR",
      "Timer + kill switch configured safely",
      "Ask us anything afterwards",
    ],
    cta: "Book setup",
    href: "mailto:admin@deemwar.com?subject=SRE%20Agent%20—%20Setup%20package",
    highlight: true,
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
            SRE Agent is free to use — bring your own API keys and run it on your
            machine or in your own cloud. It is MIT licensed, no restrictions.
            Need help getting started? A one-time setup package ($300) gets it
            running against your stack.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
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
          The Open plan is free forever — bring your own keys (Anthropic, GitHub,
          Grafana) and run it on your machine or in your own cloud. The source is
          MIT licensed on GitHub. The Setup package ($300, one-time) gets it
          running against your stack. Bigger integrations? Ask us.
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
