"use client";

type Status = "done" | "in-progress";

const sources: { name: string; category: string; status: Status; note: string }[] = [
  // Done
  { name: "Grafana", category: "Metrics & dashboards", status: "done", note: "Native data source — no bridge required." },
  { name: "Loki", category: "Logs (open-source)", status: "done", note: "Open-source observability stack as an alternative data source." },
  { name: "Prometheus", category: "Metrics (open-source)", status: "done", note: "Open-source observability stack as an alternative data source." },

  // In progress
  { name: "Cloud Logging", category: "Google Cloud logs", status: "in-progress", note: "Virtual machine + cloud logging integration coming soon." },
  { name: "Cloud Monitoring", category: "Google Cloud metrics", status: "in-progress", note: "GCP-native metric ingestion coming soon." },
  { name: "Docker", category: "Runtime targets", status: "in-progress", note: "Container-level anomaly detection coming soon." },
  { name: "Kubernetes", category: "Runtime targets", status: "in-progress", note: "Cluster + pod-level anomaly detection coming soon." },
  { name: "GitHub", category: "Source control", status: "done", note: "PR + workflow signal correlation verified." },
];

function StatusBadge({ status }: { status: Status }) {
  if (status === "done") {
    return (
      <span
        className="inline-flex items-center gap-1.5 text-[10px] font-mono uppercase tracking-[0.15em] px-2.5 py-1 rounded-full"
        style={{ backgroundColor: "rgba(63, 185, 80, 0.12)", color: "#2f7d3a" }}
      >
        <span className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: "#3fb950" }} />
        Done
      </span>
    );
  }
  return (
    <span
      className="inline-flex items-center gap-1.5 text-[10px] font-mono uppercase tracking-[0.15em] px-2.5 py-1 rounded-full"
      style={{ backgroundColor: "rgba(115, 92, 65, 0.10)", color: "#735c41" }}
    >
      <span className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: "#735c41" }} />
      In progress
    </span>
  );
}

export function DataSources() {
  return (
    <section id="data-sources" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-6xl mx-auto">
        <div className="mb-12 max-w-3xl">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            Data sources
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-4" style={{ color: "#1b1c19" }}>
            Plugs into the stack you already run.
          </h2>
          <p className="text-base sm:text-lg leading-relaxed" style={{ color: "#6b665e" }}>
            Bring your own observability backend. New integrations ship as they&apos;re
            verified — no roadmap theater.
          </p>
        </div>

        <div
          className="rounded-2xl overflow-hidden"
          style={{ border: "1px solid #e0d9cd", backgroundColor: "#ffffff" }}
        >
          <div
            className="grid grid-cols-12 gap-4 px-6 py-4 text-xs font-mono uppercase tracking-[0.15em]"
            style={{ borderBottom: "1px solid #e0d9cd", color: "#735c41", backgroundColor: "#f7f4ed" }}
          >
            <div className="col-span-5 sm:col-span-4">Source</div>
            <div className="col-span-5 sm:col-span-5">Category</div>
            <div className="col-span-2 text-right">Status</div>
          </div>

          {sources.map((s, idx) => (
            <div
              key={s.name}
              className="grid grid-cols-12 gap-4 px-6 py-5 items-center"
              style={{
                borderBottom: idx === sources.length - 1 ? "none" : "1px solid #f0ede6",
              }}
            >
              <div className="col-span-12 sm:col-span-4">
                <p className="text-base font-semibold" style={{ color: "#1b1c19" }}>
                  {s.name}
                </p>
                <p className="text-xs mt-1 sm:hidden" style={{ color: "#8b867f" }}>
                  {s.note}
                </p>
              </div>
              <div className="col-span-7 sm:col-span-5 text-sm" style={{ color: "#6b665e" }}>
                {s.category}
                <span className="hidden sm:inline">
                  <span className="mx-2" style={{ color: "#d5cfc4" }}>·</span>
                  {s.note}
                </span>
              </div>
              <div className="col-span-5 sm:col-span-3 flex sm:justify-end">
                <StatusBadge status={s.status} />
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
