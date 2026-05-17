import { useState, useCallback } from "react";
import { ReactFlowProvider } from "@xyflow/react";
import type { TraceResponse, NodeTraceResponse } from "./api/types";
import { fetchTrace } from "./api/client";
import { TraceSearch } from "./components/TraceSearch";
import { TraceSummary } from "./components/TraceSummary";
import { GraphView } from "./components/GraphView";
import { TimelineView } from "./components/TimelineView";
import { NodeDetailPanel } from "./components/NodeDetailPanel";
import { TraceSelector } from "./components/TraceSelector"; // Новый компонент

type ViewTab = "graph" | "timeline";

function App() {
  const [traces, setTraces] = useState<TraceResponse[]>([]);
  const [selectedTraceIndex, setSelectedTraceIndex] = useState<number>(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<ViewTab>("graph");

  const currentTrace = traces[selectedTraceIndex] || null;

  const handleSearch = useCallback(async (requestId: string) => {
    setLoading(true);
    setError(null);
    setTraces([]);
    setSelectedTraceIndex(0);
    setSelectedNodeId(null);
    try {
      const data = await fetchTrace(requestId);
      setTraces(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  }, []);

  const handleTraceChange = useCallback((index: number) => {
    setSelectedTraceIndex(index);
    setSelectedNodeId(null);
  }, []);

  const handleNodeClick = useCallback((nodeId: string) => {
    setSelectedNodeId(nodeId);
  }, []);

  const selectedNode: NodeTraceResponse | null =
    currentTrace?.node_traces.find((nt) => nt.node_id === selectedNodeId) ?? null;

  const hasGraphData = currentTrace?.node_traces.some(
    (nt) => nt.dependencies && nt.dependencies.length > 0
  );

  return (
    <div
      style={{
        maxWidth: 1200,
        margin: "0 auto",
        padding: "24px 16px",
        fontFamily:
          '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
      }}
    >
      <h1 style={{ fontSize: 24, marginBottom: 16 }}>Shortcut Trace Viewer</h1>

      <TraceSearch onSearch={handleSearch} loading={loading} />

      {error && (
        <div
          style={{
            marginTop: 16,
            padding: "12px 16px",
            background: "#fef2f2",
            border: "1px solid #fecaca",
            borderRadius: 8,
            color: "#dc2626",
            fontSize: 14,
          }}
        >
          {error}
        </div>
      )}

      {traces.length > 0 && currentTrace && (
        <div style={{ marginTop: 20 }}>
          {/* Селектор трейсов */}
          <TraceSelector
            traces={traces}
            selectedIndex={selectedTraceIndex}
            onSelect={handleTraceChange}
          />

          <TraceSummary trace={currentTrace} />

          <div style={{ display: "flex", gap: 8, margin: "16px 0" }}>
            {hasGraphData && (
              <TabButton
                active={activeTab === "graph"}
                onClick={() => setActiveTab("graph")}
              >
                Graph
              </TabButton>
            )}
            <TabButton
              active={activeTab === "timeline"}
              onClick={() => setActiveTab("timeline")}
            >
              Timeline
            </TabButton>
          </div>

          <div style={{ display: "flex", gap: 0 }}>
            <div style={{ flex: 1, minWidth: 0 }}>
              {activeTab === "graph" && hasGraphData ? (
                <ReactFlowProvider>
                  <GraphView
                    nodeTraces={currentTrace.node_traces}
                    onNodeClick={handleNodeClick}
                  />
                </ReactFlowProvider>
              ) : (
                <TimelineView
                  trace={currentTrace}
                  onNodeClick={handleNodeClick}
                  selectedNodeId={selectedNodeId}
                />
              )}
            </div>
            {selectedNode && (
              <NodeDetailPanel
                node={selectedNode}
                onClose={() => setSelectedNodeId(null)}
              />
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      style={{
        padding: "6px 16px",
        fontSize: 14,
        fontWeight: 500,
        border: "1px solid #d1d5db",
        borderRadius: 6,
        background: active ? "#2563eb" : "#fff",
        color: active ? "#fff" : "#374151",
        cursor: "pointer",
      }}
    >
      {children}
    </button>
  );
}

export default App;
