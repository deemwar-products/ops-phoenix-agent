"use client";

export function Problem() {
  return (
    <section id="problem" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-6xl mx-auto">
        <div className="max-w-3xl">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            The problem
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-6" style={{ color: "#1b1c19" }}>
            Nobody is watching what an incident costs — per minute, per engineer, per page.
          </h2>
          <p className="text-base sm:text-lg leading-relaxed mb-12" style={{ color: "#6b665e" }}>
            Once an alert fires, the real timer starts. Engineers scramble across
            dashboards, grep through logs, cross-reference deploys, and build a
            timeline — all before writing a single line of fix. A single incident
            can consume hours of senior engineer time. And the next one is waiting.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {[
            {
              label: "Alert fatigue",
              stat: "50+ alerts/day",
              detail: "Engineers learn to ignore the bell — until the real one gets buried.",
            },
            {
              label: "Investigation time",
              stat: "30+ min",
              detail: "Same five dashboards, same log grep, different day.",
            },
            {
              label: "2 AM pages",
              stat: "Avoidable",
              detail: "Issues that could have been fixed — or not paged at all — before anyone noticed.",
            },
            {
              label: "Repeat incidents",
              stat: "Same root cause",
              detail: "The fix is known, the runbook exists. It's just never automated.",
            },
          ].map((pain) => (
            <div
              key={pain.label}
              className="p-6 rounded-xl"
              style={{ backgroundColor: "#f5f2ec", border: "1px solid #e8e3d9" }}
            >
              <p className="text-xs font-semibold uppercase tracking-[0.15em] mb-2 font-mono" style={{ color: "#735c41" }}>
                {pain.label}
              </p>
              <p className="text-2xl font-bold mb-2" style={{ color: "#1b1c19" }}>{pain.stat}</p>
              <p className="text-sm leading-relaxed" style={{ color: "#6b665e" }}>{pain.detail}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
