import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  type NodeTypes,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { useMemo } from "react";
import type { NodeTraceResponse } from "../api/types";
import { buildLayout } from "../lib/layout";
import { GraphNode } from "./GraphNode";

const nodeTypes: NodeTypes = {
  graphNode: GraphNode,
};

interface Props {
  nodeTraces: NodeTraceResponse[];
  onNodeClick: (nodeId: string) => void;
}

export function GraphView({ nodeTraces, onNodeClick }: Props) {
  const { nodes, edges } = useMemo(
    () => buildLayout(nodeTraces),
    [nodeTraces]
  );

  return (
    <div style={{ width: "100%", height: 500 }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodeClick={(_, node) => onNodeClick(node.id)}
        fitView
        proOptions={{ hideAttribution: true }}
      >
        <Background />
        <Controls />
        <MiniMap />
      </ReactFlow>
    </div>
  );
}
