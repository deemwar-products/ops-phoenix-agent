"use client";

const pains = [
  {
    label: "Alert fatigue",
    stat: "50+ alerts/day",
    detail: "Engineers mute the bell. Then the real one gets buried in the noise.",
  },
  {
    label: "Investigation time",
    stat: "30+ min",
    detail: "Same dashboards, same log grep, different day. Senior engineers doing junior detective work.",
  },
  {
    label: "2 AM pages",
    stat: "Avoidable",
    detail: "Issues that could have been fixed before anyone noticed. While you sleep.",
  },
  {
    label: "Repeat incidents",
    stat: "Known root cause",
    detail: "The fix is known, the runbook exists. It's just never automated.",
  },
];

export function Problem() {
  return (
    <section id="problem" className="py-32 px-6" style={{ backgroundColor: "#0f1117" }}>
      <div className="max-w-6xl mx-auto">
        <div className="max-w-3xl">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            The problem
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-6" style={{ color: "#f0ede6" }}>
            Your on-call engineer is the bottleneck.
          </h2>
          <p className="text-base sm:text-lg leading-relaxed mb-12" style={{ color: "#8b867f" }}>
            Once an alert fires, the real cost begins. Minutes stretch into hours
            of dashboard-hopping and log grepping. Meanwhile, your users are
            staring at error pages and your revenue is bleeding.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {pains.map((pain) => (
            <div
              key={pain.label}
              className="p-6 rounded-xl"
              style={{ backgroundColor: "rgba(255,255,255,0.03)", border: "1px solid rgba(255,255,255,0.06)" }}
            >
              <p className="text-xs font-semibold uppercase tracking-[0.15em] mb-2 font-mono" style={{ color: "#735c41" }}>
                {pain.label}
              </p>
              <p className="text-2xl font-bold mb-2" style={{ color: "#f0ede6" }}>{pain.stat}</p>
              <p className="text-sm leading-relaxed" style={{ color: "#8b867f" }}>{pain.detail}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
