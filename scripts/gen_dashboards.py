#!/usr/bin/env python3
import argparse
import json
import os
import re
import sys
import tempfile
from dataclasses import dataclass, field
from pathlib import Path
from typing import Iterable

import yaml


HTTP_ROUTER_FILE = "http_router.yaml"
GRAPHS_DIR = "graphs"
DEFAULT_INPUT_NODE = "input"

PANELS_PER_ROW = 24
PANEL_HEIGHT = 8
GRAPH_PANEL_WIDTH = 8


@dataclass(frozen=True)
class Route:
    name: str
    path: str
    method: str


@dataclass
class NodeInfo:
    id: str
    cache_enabled: bool = False


@dataclass
class Graph:
    namespace_id: str
    graph_id: str
    nodes: list[NodeInfo]
    routes: tuple[Route, ...] = field(default_factory=tuple)

    @property
    def node_ids(self) -> tuple[str, ...]:
        return tuple(n.id for n in self.nodes)

    @property
    def title(self) -> str:
        return f"{self.namespace_id} / {self.graph_id}"

    @property
    def slug(self) -> str:
        return re.sub(r"[^a-z0-9_-]+", "-", f"{self.namespace_id}-{self.graph_id}".lower()).strip("-")


def load_routes(namespace_dir: Path) -> dict[str, list[Route]]:
    router_path = namespace_dir / HTTP_ROUTER_FILE
    if not router_path.is_file():
        return {}
    with router_path.open() as fh:
        raw = yaml.safe_load(fh) or {}
    out: dict[str, list[Route]] = {}
    for name, route in (raw.get("routes") or {}).items():
        graph_id = route.get("graph")
        path = route.get("path", "")
        method = (route.get("method") or "GET").upper()
        if not graph_id or not path:
            continue
        out.setdefault(graph_id, []).append(Route(name=name, path=path, method=method))
    for graph_id in out:
        out[graph_id].sort(key=lambda r: (r.method, r.path, r.name))
    return out


def load_graph_nodes(graph_file: Path) -> list[NodeInfo]:
    with graph_file.open() as fh:
        raw = yaml.safe_load(fh) or {}
    
    nodes_data = raw.get("nodes") or {}
    nodes = []
    
    for node_id, node_config in nodes_data.items():
        cache_enabled = node_config.get("cache", {}).get("enabled", False)
        nodes.append(NodeInfo(id=node_id, cache_enabled=cache_enabled))
    
    input_node = raw.get("input-node") or DEFAULT_INPUT_NODE
    if input_node not in [n.id for n in nodes]:
        nodes.append(NodeInfo(id=input_node, cache_enabled=False))
    
    return sorted(nodes, key=lambda n: n.id)


def discover_graphs(configs_dir: Path) -> list[Graph]:
    out: list[Graph] = []
    for namespace_dir in sorted(p for p in configs_dir.iterdir() if p.is_dir()):
        routes_by_graph = load_routes(namespace_dir)
        graphs_dir = namespace_dir / GRAPHS_DIR
        if not graphs_dir.is_dir():
            continue
        for graph_file in sorted(graphs_dir.glob("*.yaml")):
            graph_id = graph_file.stem
            nodes = load_graph_nodes(graph_file)
            routes = tuple(routes_by_graph.get(graph_id, ()))
            out.append(Graph(
                namespace_id=namespace_dir.name,
                graph_id=graph_id,
                nodes=nodes,
                routes=routes,
            ))
    return out


def regex_alt(values: Iterable[str]) -> str:
    escaped = []
    for v in values:
        special = r'.*+?^${}[]()|\\'
        for char in special:
            if char in v:
                v = v.replace(char, "\\" + char)
        escaped.append(v)
    
    escaped = sorted(escaped)
    
    if len(escaped) == 1:
        return escaped[0]
    
    return "(" + "|".join(escaped) + ")"


def grid_pos(index: int, *, y_base: int, width: int = GRAPH_PANEL_WIDTH, height: int = PANEL_HEIGHT) -> dict:
    per_row = PANELS_PER_ROW // width
    return {
        "x": (index % per_row) * width,
        "y": y_base + (index // per_row) * height,
        "w": width,
        "h": height,
    }


def panel_timeseries(*, panel_id: int, title: str, datasource: str, queries: list[dict], unit: str | None = None, grid: dict, decimals: int | None = None) -> dict:
    field_config = {
        "defaults": {
            "custom": {
                "drawStyle": "line",
                "fillOpacity": 10,
                "lineWidth": 1,
            },
        },
        "overrides": [],
    }
    if unit:
        field_config["defaults"]["unit"] = unit
    if decimals is not None:
        field_config["defaults"]["decimals"] = decimals
    return {
        "id": panel_id,
        "type": "timeseries",
        "title": title,
        "datasource": {"type": "prometheus", "uid": datasource},
        "gridPos": grid,
        "targets": queries,
        "fieldConfig": field_config,
        "options": {
            "legend": {"displayMode": "list", "placement": "bottom", "showLegend": True},
            "tooltip": {"mode": "multi", "sort": "desc"},
        },
    }


def target(expr: str, *, ref_id: str, legend: str) -> dict:
    return {
        "refId": ref_id,
        "expr": expr,
        "legendFormat": legend,
        "datasource": {"type": "prometheus", "uid": "${datasource}"},
    }


def cluster_resources_row(*, datasource: str, panel_id_start: int) -> tuple[dict, list[dict], int]:
    row = {
        "id": panel_id_start,
        "type": "row",
        "title": "Cluster resources",
        "collapsed": False,
        "gridPos": {"x": 0, "y": 0, "w": PANELS_PER_ROW, "h": 1},
        "panels": [],
    }
    cpu_expr = (
        'sum by(pod) (rate(container_cpu_usage_seconds_total{'
        'namespace=~"$k8s_namespace",pod=~"$pod",container!="",container!="POD"'
        '}[$__rate_interval]))'
    )
    mem_expr = (
        'sum by(pod) (container_memory_working_set_bytes{'
        'namespace=~"$k8s_namespace",pod=~"$pod",container!="",container!="POD"'
        '})'
    )
    cpu_panel = panel_timeseries(
        panel_id=panel_id_start + 1,
        title="CPU usage by pod (cores)",
        datasource=datasource,
        queries=[target(cpu_expr, ref_id="A", legend="{{pod}}")],
        unit="none",
        grid={"x": 0, "y": 1, "w": 12, "h": PANEL_HEIGHT},
    )
    mem_panel = panel_timeseries(
        panel_id=panel_id_start + 2,
        title="Memory by pod",
        datasource=datasource,
        queries=[target(mem_expr, ref_id="A", legend="{{pod}}")],
        unit="bytes",
        grid={"x": 12, "y": 1, "w": 12, "h": PANEL_HEIGHT},
    )
    return row, [cpu_panel, mem_panel], panel_id_start + 3


def add_node_section(*, panels: list, node: NodeInfo, graph: Graph, datasource: str, service: str, next_id: int, y_base: int) -> tuple[int, int]:
    current_y = y_base
    
    node_header = {
        "id": next_id,
        "type": "text",
        "title": f"Node: {node.id}",
        "collapsed": False,
        "gridPos": {"x": 0, "y": current_y, "w": PANELS_PER_ROW, "h": 1},
        "panels": [],
    }
    panels.append(node_header)
    next_id += 1
    current_y += 1
    panel_index = 0
    
    node_common = f'service="{service}",namespace="{graph.namespace_id}",graph_id="{graph.graph_id}",node_id="{node.id}"'
    
    node_rps = f'rate(shortcut_node_requests_total{{{node_common}}}[$__rate_interval])'
    panels.append(panel_timeseries(
        panel_id=next_id,
        title=f"RPS",
        datasource=datasource,
        queries=[target(node_rps, ref_id="A", legend="RPS")],
        unit="reqps",
        grid=grid_pos(panel_index, y_base=current_y, width=8, height=PANEL_HEIGHT),
    ))
    next_id += 1
    panel_index += 1
    
    node_lat_50 = f'shortcut_node_duration_seconds{{{node_common},quantile="0.5"}}'
    node_lat_90 = f'shortcut_node_duration_seconds{{{node_common},quantile="0.9"}}'
    node_lat_95 = f'shortcut_node_duration_seconds{{{node_common},quantile="0.95"}}'
    node_lat_99 = f'shortcut_node_duration_seconds{{{node_common},quantile="0.99"}}'
    
    panels.append(panel_timeseries(
        panel_id=next_id,
        title=f"Latency",
        datasource=datasource,
        queries=[
            target(node_lat_50, ref_id="A", legend="p50"),
            target(node_lat_90, ref_id="B", legend="p90"),
            target(node_lat_95, ref_id="C", legend="p95"),
            target(node_lat_99, ref_id="D", legend="p99"),
        ],
        unit="s",
        grid=grid_pos(panel_index, y_base=current_y, width=8, height=PANEL_HEIGHT),
    ))
    next_id += 1
    panel_index += 1
    
    node_err = (
        f'rate(shortcut_node_errors_total{{{node_common}}}[$__rate_interval])'
        ' / clamp_min('
        f'rate(shortcut_node_requests_total{{{node_common}}}[$__rate_interval])'
        ', 1e-9)'
        ' or on() vector(0)'
    )

    errors_panel = panel_timeseries(
        panel_id=next_id,
        title=f"Error Ratio",
        datasource=datasource,
        queries=[target(node_err, ref_id="A", legend="Error Ratio")],
        unit="percentunit",
        grid=grid_pos(panel_index, y_base=current_y, width=8, height=PANEL_HEIGHT),
        decimals=3,
    )

    errors_panel["fieldConfig"]["defaults"]["max"] = 1.0
    errors_panel["fieldConfig"]["defaults"]["min"] = 0.0

    panels.append(errors_panel)
    next_id += 1
    
    current_y += PANEL_HEIGHT
    
    if node.cache_enabled:
        cache_header = {
            "id": next_id,
            "type": "text",
            "title": f"Cache: {node.id}",
            "collapsed": False,
            "gridPos": {"x": 0, "y": current_y, "w": PANELS_PER_ROW, "h": 1},
            "panels": [],
        }
        panels.append(cache_header)
        next_id += 1
        current_y += 1
        
        cache_common = f'node_id="{node.id}"'
        
        inserts_expr = f'rate(shortcut_cache_inserts_total{{{cache_common}}}[$__rate_interval])'
        panels.append(panel_timeseries(
            panel_id=next_id,
            title=f"Inserts Rate",
            datasource=datasource,
            queries=[target(inserts_expr, ref_id="A", legend="Inserts/s")],
            unit="ops",
            grid={"x": 0, "y": current_y, "w": 12, "h": PANEL_HEIGHT},
        ))
        next_id += 1
        
        hits_expr = f'rate(shortcut_cache_hits_total{{{cache_common}}}[$__rate_interval])'
        misses_expr = f'rate(shortcut_cache_misses_total{{{cache_common}}}[$__rate_interval])'
        panels.append(panel_timeseries(
            panel_id=next_id,
            title=f"Hits vs Misses",
            datasource=datasource,
            queries=[
                target(hits_expr, ref_id="A", legend="Hits/s"),
                target(misses_expr, ref_id="B", legend="Misses/s"),
            ],
            unit="ops",
            grid={"x": 12, "y": current_y, "w": 12, "h": PANEL_HEIGHT},
        ))
        next_id += 1
        
        current_y += PANEL_HEIGHT
    
    return next_id, current_y


def graph_row(*, graph: Graph, datasource: str, service: str, panel_id_start: int, y_base: int) -> tuple[dict, int]:
    panels: list[dict] = []
    next_id = panel_id_start + 1
    inner_y = y_base + 1

    if graph.routes:
        path_re = regex_alt(r.path for r in graph.routes)
        method_re = regex_alt(r.method for r in graph.routes)
        
        common = f'service="{service}",namespace="{graph.namespace_id}",path=~"{path_re}",method=~"{method_re}"'

        rps_expr = (
            f'sum by(path,method) (rate(http_requests_total{{{common}}}[$__rate_interval]))'
        )
        panels.append(panel_timeseries(
            panel_id=next_id,
            title=f"HTTP RPS",
            datasource=datasource,
            queries=[target(rps_expr, ref_id="A", legend="{{method}} {{path}}")],
            unit="reqps",
            grid=grid_pos(0, y_base=inner_y, width=8, height=PANEL_HEIGHT),
        ))
        next_id += 1

        lat_expr = (
            f'http_request_duration_quantiles_seconds{{{common},quantile="0.95"}}'
        )
        panels.append(panel_timeseries(
            panel_id=next_id,
            title=f"HTTP latency",
            datasource=datasource,
            queries=[target(lat_expr, ref_id="A", legend="{{method}} {{path}}")],
            unit="s",
            grid=grid_pos(1, y_base=inner_y, width=8, height=PANEL_HEIGHT),
        ))
        next_id += 1

        err_expr = (
            f'sum by(path,method) (rate(http_codes_total{{{common},code=~"4..|5.."}}[$__rate_interval]))'
            ' / clamp_min('
            f'sum by(path,method) (rate(http_codes_total{{{common}}}[$__rate_interval]))'
            ', 1e-9)'
            ' or on() vector(0)'
        )

        com_errors_panel = panel_timeseries(
            panel_id=next_id,
            title=f"HTTP error ratio",
            datasource=datasource,
            queries=[target(err_expr, ref_id="A", legend="{{method}} {{path}}")],
            unit="percentunit",
            grid=grid_pos(2, y_base=inner_y, width=8, height=PANEL_HEIGHT),
            decimals=3,
        )

        com_errors_panel["fieldConfig"]["defaults"]["max"] = 1.0
        com_errors_panel["fieldConfig"]["defaults"]["min"] = 0.0

        panels.append(com_errors_panel)
        next_id += 1
        
        inner_y += PANEL_HEIGHT

    if graph.nodes:
        node_common_all = f'service="{service}",namespace="{graph.namespace_id}",graph_id="{graph.graph_id}"'
        
        all_nodes_rps = f'sum by(node_id) (rate(shortcut_node_requests_total{{{node_common_all}}}[$__rate_interval]))'
        panels.append(panel_timeseries(
            panel_id=next_id,
            title=f"All Nodes RPS",
            datasource=datasource,
            queries=[target(all_nodes_rps, ref_id="A", legend="{{node_id}}")],
            unit="reqps",
            grid=grid_pos(0, y_base=inner_y, width=8, height=PANEL_HEIGHT),
        ))
        next_id += 1
        
        all_nodes_lat = f'shortcut_node_duration_seconds{{{node_common_all},quantile="0.95"}}'
        panels.append(panel_timeseries(
            panel_id=next_id,
            title=f"All Nodes Latency",
            datasource=datasource,
            queries=[target(all_nodes_lat, ref_id="A", legend="{{node_id}}")],
            unit="s",
            grid=grid_pos(1, y_base=inner_y, width=8, height=PANEL_HEIGHT),
        ))
        next_id += 1
        
        all_nodes_err = (
            f'sum by(node_id) (rate(shortcut_node_errors_total{{{node_common_all}}}[$__rate_interval]))'
            ' / clamp_min('
            f'sum by(node_id) (rate(shortcut_node_requests_total{{{node_common_all}}}[$__rate_interval]))'
            ', 1e-9)'
            ' or on() vector(0)'
        )

        all_errors_panel = panel_timeseries(
            panel_id=next_id,
            title=f"All Nodes Error Ratio",
            datasource=datasource,
            queries=[target(all_nodes_err, ref_id="A", legend="{{node_id}}")],
            unit="percentunit",
            grid=grid_pos(2, y_base=inner_y, width=8, height=PANEL_HEIGHT),
            decimals=3,
        )

        all_errors_panel["fieldConfig"]["defaults"]["max"] = 1.0
        all_errors_panel["fieldConfig"]["defaults"]["min"] = 0.0

        panels.append(all_errors_panel)
        next_id += 1
        
        inner_y += PANEL_HEIGHT
        
        for node in graph.nodes:
            next_id, inner_y = add_node_section(
                panels=panels,
                node=node,
                graph=graph,
                datasource=datasource,
                service=service,
                next_id=next_id,
                y_base=inner_y
            )

    row = {
        "id": panel_id_start,
        "type": "row",
        "title": graph.title,
        "collapsed": True,
        "gridPos": {"x": 0, "y": y_base, "w": PANELS_PER_ROW, "h": 1},
        "panels": panels,
    }
    return row, next_id


def build_dashboard(graphs: list[Graph], *, datasource: str, service: str) -> dict:
    panels: list[dict] = []
    panel_id = 1

    cluster_row, cluster_panels, panel_id = cluster_resources_row(datasource=datasource, panel_id_start=panel_id)
    cluster_row["panels"] = cluster_panels
    panels.append(cluster_row)

    y_cursor = 1 + PANEL_HEIGHT
    for graph in sorted(graphs, key=lambda g: (g.namespace_id, g.graph_id)):
        row, panel_id = graph_row(
            graph=graph,
            datasource=datasource,
            service=service,
            panel_id_start=panel_id,
            y_base=y_cursor,
        )
        panels.append(row)
        y_cursor += 1

    templating = {
        "list": [
            {
                "name": "datasource",
                "type": "datasource",
                "query": "prometheus",
                "current": {"text": "Prometheus", "value": "Prometheus"},
                "hide": 0,
                "label": "Datasource",
                "refresh": 1,
            },
            {
                "name": "namespace",
                "type": "query",
                "datasource": {"type": "prometheus", "uid": "${datasource}"},
                "query": "label_values(http_requests_total, namespace)",
                "refresh": 2,
                "includeAll": True,
                "allValue": ".*",
                "current": {"text": "All", "value": "$__all"},
                "hide": 0,
                "label": "Namespace",
                "sort": 1,
            },
            {
                "name": "k8s_namespace",
                "type": "query",
                "datasource": {"type": "prometheus", "uid": "${datasource}"},
                "query": "label_values(kube_pod_info, namespace)",
                "refresh": 2,
                "includeAll": True,
                "allValue": ".*",
                "current": {"text": "All", "value": "$__all"},
                "hide": 0,
                "label": "k8s namespace",
                "sort": 1,
            },
            {
                "name": "pod",
                "type": "query",
                "datasource": {"type": "prometheus", "uid": "${datasource}"},
                "query": 'label_values(kube_pod_info{namespace=~"$k8s_namespace"}, pod)',
                "refresh": 2,
                "includeAll": True,
                "allValue": ".*",
                "current": {"text": "All", "value": "$__all"},
                "hide": 0,
                "label": "Pod",
                "sort": 1,
            },
            {
                "name": "path",
                "type": "query",
                "datasource": {"type": "prometheus", "uid": "${datasource}"},
                "query": f'label_values(http_requests_total{{namespace=~"$namespace"}}, path)',
                "refresh": 2,
                "includeAll": True,
                "allValue": ".*",
                "current": {"text": "All", "value": "$__all"},
                "hide": 0,
                "label": "Path",
                "sort": 1,
            },
        ],
    }

    return {
        "annotations": {"list": []},
        "editable": True,
        "fiscalYearStartMonth": 0,
        "graphTooltip": 0,
        "links": [],
        "liveNow": False,
        "panels": panels,
        "refresh": "30s",
        "schemaVersion": 39,
        "tags": ["shortcut", "generated"],
        "templating": templating,
        "time": {"from": "now-1h", "to": "now"},
        "timepicker": {},
        "timezone": "",
        "title": "Shortcut",
        "uid": "shortcut-overview",
        "weekStart": "",
    }


def write_json(path: Path, data: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(
        "w", delete=False, dir=path.parent, prefix=path.name + ".", suffix=".tmp"
    ) as tmp:
        json.dump(data, tmp, indent=2, sort_keys=False)
        tmp.write("\n")
        tmp_path = Path(tmp.name)
    os.replace(tmp_path, path)


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description="Generate Grafana dashboards from shortcut graph configs")
    parser.add_argument("--configs-dir", required=True, type=Path, help="Path to a directory with namespace subdirectories")
    parser.add_argument("--out-dir", type=Path, default=Path("k8s/dashboards"), help="Output directory for dashboard JSON files")
    parser.add_argument("--datasource", default="Prometheus", help="Prometheus datasource UID variable default")
    parser.add_argument("--service", default="shortcut", help="Value of the `service` label written by shortcut")
    args = parser.parse_args(argv)

    configs_dir: Path = args.configs_dir
    if not configs_dir.is_dir():
        print(f"configs dir not found: {configs_dir}", file=sys.stderr)
        return 1

    graphs = discover_graphs(configs_dir)
    if not graphs:
        print(f"no graphs discovered under {configs_dir}", file=sys.stderr)
        return 1

    dashboard = build_dashboard(graphs, datasource=args.datasource, service=args.service)
    out_path: Path = args.out_dir / "shortcut.json"
    write_json(out_path, dashboard)
    print(f"wrote {out_path} (graphs: {len(graphs)})")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))