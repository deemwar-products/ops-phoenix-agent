"use client";

import { ContactModal } from "./ContactModal";

const plans = [
  {
    name: "Open Source",
    price: "Free",
    period: "forever",
    description: "MIT licensed. Bring your own keys. Run it on your machine or in your own cloud.",
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
            Free to use. Open source.
          </h2>
          <p className="text-base" style={{ color: "#6b665e" }}>
            SRE Agent is free to use — bring your own API keys and run it on your
            machine or in your own cloud. It is MIT licensed, no restrictions, no
            paid tiers. Need a hand getting it running? Reach out and we'll help.
          </p>
        </div>

        <div className="grid grid-cols-1 max-w-lg mx-auto">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className="rounded-2xl overflow-hidden"
              style={{
                border: "1px solid #e0d9cd",
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
              </div>
            </div>
          ))}
        </div>

        <p className="mt-10 text-sm" style={{ color: "#8b867f" }}>
          SRE Agent is free forever — bring your own keys (Anthropic, GitHub,
          Grafana) and run it on your machine or in your own cloud. The source is
          MIT licensed on GitHub. No paid tiers, no seat limits.
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
