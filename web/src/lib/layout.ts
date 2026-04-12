import dagre from "@dagrejs/dagre";
import type { Node, Edge } from "@xyflow/react";
import type { NodeTraceResponse } from "../api/types";

export interface GraphNodeData extends Record<string, unknown> {
  label: string;
  nodeType: string;
  durationMs: number;
  statusCode: number;
  retryCount: number;
  cached: boolean;
  error: string;
  status: "ok" | "error" | "retried" | "cached";
}

const NODE_WIDTH = 180;
const NODE_HEIGHT = 60;

export function buildLayout(nodeTraces: NodeTraceResponse[]): {
  nodes: Node<GraphNodeData>[];
  edges: Edge[];
} {
  const g = new dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({ rankdir: "TB", nodesep: 60, ranksep: 80 });

  for (const nt of nodeTraces) {
    g.setNode(nt.node_id, { width: NODE_WIDTH, height: NODE_HEIGHT });
  }

  const edges: Edge[] = [];
  for (const nt of nodeTraces) {
    if (nt.dependencies) {
      for (const dep of nt.dependencies) {
        const edgeId = `${dep.node_id}->${nt.node_id}`;
        g.setEdge(dep.node_id, nt.node_id);
        edges.push({
          id: edgeId,
          source: dep.node_id,
          target: nt.node_id,
          animated: true,
        });
      }
    }
  }

  dagre.layout(g);

  const nodes: Node<GraphNodeData>[] = nodeTraces.map((nt) => {
    const pos = g.node(nt.node_id);
    const hasError = !!nt.error;
    const wasCached = nt.cached ?? false;
    const wasRetried = (nt.retry_count ?? 0) > 0;

    return {
      id: nt.node_id,
      position: {
        x: (pos?.x ?? 0) - NODE_WIDTH / 2,
        y: (pos?.y ?? 0) - NODE_HEIGHT / 2,
      },
      data: {
        label: nt.node_id,
        nodeType: nt.node_type ?? "unknown",
        durationMs: nt.duration_ms,
        statusCode: nt.status_code ?? 0,
        retryCount: nt.retry_count ?? 0,
        cached: wasCached,
        error: nt.error ?? "",
        status: hasError ? "error" : wasCached ? "cached" : wasRetried ? "retried" : "ok",
      },
      type: "graphNode",
    };
  });

  return { nodes, edges };
}
