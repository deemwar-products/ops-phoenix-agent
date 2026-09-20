"use client";

const plans = [
  {
    name: "Individual",
    description:
      "One engineer, one service. Guided mode with manual approval on every step.",
    features: [
      "1 service / repo",
      "Anomaly detection + RCA",
      "PR-based fixes (manual approval)",
      "Local run history",
      "Email support",
    ],
    cta: "Talk to us",
  },
  {
    name: "Team",
    description:
      "Shared on-call rotation. Guided or autonomous fixes with shared run history.",
    features: [
      "Up to 10 services / repos",
      "Anomaly detection + RCA",
      "Guided or autonomous mode",
      "Slack integration",
      "Shared run history + audit log",
    ],
    cta: "Talk to us",
  },
  {
    name: "Enterprise",
    description:
      "Full autonomy, multi-cloud, SSO, and dedicated support with SLA.",
    features: [
      "Unlimited services / repos",
      "Autonomous mode (no approvals)",
      "Multi-cloud + on-prem",
      "SSO, RBAC, audit logs",
      "Custom integrations",
      "Dedicated support + SLA",
    ],
    cta: "Talk to us",
  },
];

function checkFeature(planName: string, feature: string): boolean {
  const map: Record<string, string[]> = {
    "1 service / repo": ["Individual"],
    "Up to 10 services / repos": ["Team", "Enterprise"],
    "Unlimited services / repos": ["Enterprise"],
    "Anomaly detection + RCA": ["Individual", "Team", "Enterprise"],
    "PR-based fixes (manual approval)": ["Individual"],
    "Guided or autonomous mode": ["Team", "Enterprise"],
    "Autonomous mode (no approvals)": ["Enterprise"],
    "Slack integration": ["Team", "Enterprise"],
    "Shared run history + audit log": ["Team", "Enterprise"],
    "Email support": ["Individual", "Team"],
    "Multi-cloud + on-prem": ["Enterprise"],
    "SSO, RBAC, audit logs": ["Enterprise"],
    "Custom integrations": ["Enterprise"],
    "Dedicated support + SLA": ["Enterprise"],
  };
  return (map[feature] || []).includes(planName);
}

const allFeatures = [
  "1 service / repo",
  "Up to 10 services / repos",
  "Unlimited services / repos",
  "Anomaly detection + RCA",
  "PR-based fixes (manual approval)",
  "Guided or autonomous mode",
  "Autonomous mode (no approvals)",
  "Slack integration",
  "Shared run history + audit log",
  "Email support",
  "Multi-cloud + on-prem",
  "SSO, RBAC, audit logs",
  "Custom integrations",
  "Dedicated support + SLA",
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
            Plans for every team size.
          </h2>
          <p className="text-base" style={{ color: "#6b665e" }}>
            Pricing will be announced soon. Reach out to get on the early access list.
          </p>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-sm" style={{ color: "#1b1c19" }}>
            <thead>
              <tr style={{ borderBottom: "2px solid #e0d9cd" }}>
                <th className="text-left py-4 pr-8 font-medium text-xs uppercase tracking-[0.15em]" style={{ color: "#735c41", verticalAlign: "middle" }}>
                  What&apos;s included
                </th>
                {plans.map((plan) => (
                  <th
                    key={plan.name}
                    className="text-left py-4 px-6 font-medium text-xs uppercase tracking-[0.15em]"
                    style={{ color: "#735c41", verticalAlign: "middle" }}
                  >
                    {plan.name}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {allFeatures.map((feature) => (
                <tr key={feature} style={{ borderBottom: "1px solid #f0ede6" }}>
                  <td className="py-3.5 pr-8" style={{ color: "#6b665e", verticalAlign: "middle" }}>
                    {feature}
                  </td>
                  {plans.map((plan) => (
                    <td key={plan.name} className="py-3.5 px-6 text-center" style={{ verticalAlign: "middle" }}>
                      {checkFeature(plan.name, feature) ? (
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" className="inline-block" style={{ color: "#735c41" }}>
                          <path d="M5 13l4 4L19 7" />
                        </svg>
                      ) : (
                        <span style={{ color: "#d5cfc4" }}>—</span>
                      )}
                    </td>
                  ))}
                </tr>
              ))}
              <tr>
                <td className="py-5 pr-8" style={{ verticalAlign: "middle" }} />
                {plans.map((plan) => (
                  <td key={plan.name} className="py-5 px-6" style={{ verticalAlign: "middle" }}>
                    <a
                      href="https://deemwar.com/contact"
                      className="inline-block px-6 py-2.5 rounded-full text-sm font-semibold transition-colors"
                      style={{
                        backgroundColor: plan.name === "Team" ? "#735c41" : "transparent",
                        color: plan.name === "Team" ? "#fff" : "#735c41",
                        border: plan.name === "Team" ? "none" : "1px solid #d5cfc4",
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
      </div>
    </section>
  );
}
