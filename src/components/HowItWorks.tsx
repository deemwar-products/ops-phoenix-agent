"use client";

const steps = [
  {
    num: "i.",
    title: "Detect",
    body:
      "Pulls metrics from your Grafana / Cloud Monitoring backend. Learns your traffic patterns and flags deviations before they trigger cascading failures.",
  },
  {
    num: "ii.",
    title: "Analyze",
    body:
      "Reads log entries, correlates with recent deploys and traces, and identifies the likely root cause. Returns a confidence score with the finding.",
  },
  {
    num: "iii.",
    title: "Fix",
    body:
      "Generates a real code diff targeting the root cause. Opens a PR with the fix — but nothing deploys until a human reviews and approves it. You stay in control.",
  },
  {
    num: "iv.",
    title: "Verify",
    body:
      "Monitors key metrics post-deploy. If the fix worked, closes the incident. If not, rolls back automatically. Closes the loop — not just the ticket.",
  },
];

export function HowItWorks() {
  return (
    <section id="how-it-works" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-6xl mx-auto">
        <div className="mb-16 max-w-3xl">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            How it works
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-4" style={{ color: "#1b1c19" }}>
            Every request takes the same four steps, locally.
          </h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-x-16 gap-y-12">
          {steps.map((s) => (
            <div key={s.title}>
              <div className="flex items-baseline gap-3 mb-3">
                <span className="font-mono text-sm" style={{ color: "#735c41" }}>{s.num}</span>
                <h3 className="text-xl font-bold" style={{ color: "#1b1c19" }}>{s.title}</h3>
              </div>
              <p className="text-sm leading-relaxed pl-7" style={{ color: "#6b665e" }}>
                {s.body}
              </p>
            </div>
          ))}
        </div>

        <div
          className="mt-20 p-8 rounded-2xl"
          style={{ backgroundColor: "#f0ede6", border: "1px solid #e0d9cd" }}
        >
          <p className="text-xs font-semibold uppercase tracking-[0.2em] mb-2 font-mono" style={{ color: "#735c41" }}>
            TOTAL TIME
          </p>
          <p className="text-4xl font-bold mb-2" style={{ color: "#1b1c19" }}>&lt; 5 minutes</p>
          <p className="text-sm" style={{ color: "#6b665e" }}>
            From first anomaly signal to verified fix. Zero human intervention required
            (if you want it).
          </p>
        </div>
      </div>
    </section>
  );
}
