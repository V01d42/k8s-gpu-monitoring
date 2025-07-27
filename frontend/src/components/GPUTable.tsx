import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import TableSortLabel from "@mui/material/TableSortLabel";
import * as React from "react";

import type { GpuRow } from "../types/gpu";

import type { ApiResponse, GPUMetrics } from "../types/api";
import { mockGpuMetrics } from "../types/api.mock";
import { getConfig } from "../utils/config";

const config = getConfig();
const API_BASE_URL = config.API_BASE_URL;
/**
 * Fetch GPU metrics from the API. If the request fails, return mock data.
 */
async function fetchGpuMetricsWithFallback(): Promise<
  ApiResponse<GPUMetrics[]>
> {
  try {
    console.log(
      `Fetching GPU metrics from: ${API_BASE_URL}/api/v1/gpu/metrics`
    );
    const res = await fetch(`${API_BASE_URL}/api/v1/gpu/metrics`);
    if (!res.ok) throw new Error(`API error: ${res.status}`);
    const data = (await res.json()) as ApiResponse<GPUMetrics[]>;
    return data;
  } catch (e) {
    console.log("Failed to fetch GPU metrics:", e);
    console.log("use mock data due to fetch error");
    return mockGpuMetrics;
  }
}

type Order = "asc" | "desc";
type GpuRowKey = keyof GpuRow;

const columns: {
  id: GpuRowKey;
  label: string;
  align?: "right" | "center" | "left";
}[] = [
  { id: "node_name", label: "node_name", align: "right" },
  { id: "timestamp", label: "timestamp", align: "right" },
  { id: "gpu_index", label: "gpu_index", align: "right" },
  { id: "gpu_name", label: "gpu_name", align: "right" },
  { id: "utilization", label: "utilization (%)", align: "right" },
  { id: "memory_used", label: "memory_used (MiB)", align: "right" },
  { id: "memory_total", label: "memory_total (MiB)", align: "right" },
  { id: "memory_utilization", label: "memory_utilization (%)", align: "right" },
  { id: "temperature", label: "temperature (°C)", align: "right" },
  { id: "power_draw", label: "power_draw (W)", align: "right" },
  { id: "power_limit", label: "power_limit (W)", align: "right" },
];

function descendingComparator<T>(a: T, b: T, orderBy: keyof T) {
  const aValue = a[orderBy];
  const bValue = b[orderBy];
  if (typeof aValue === "number" && typeof bValue === "number") {
    return bValue - aValue;
  }
  if (typeof aValue === "string" && typeof bValue === "string") {
    return bValue.localeCompare(aValue);
  }
  return 0;
}

function getComparator<T>(order: Order, orderBy: keyof T) {
  return order === "desc"
    ? (a: T, b: T) => descendingComparator(a, b, orderBy)
    : (a: T, b: T) => -descendingComparator(a, b, orderBy);
}

export default function GPUTable() {
  const [order, setOrder] = React.useState<Order>("asc");
  const [orderBy, setOrderBy] = React.useState<GpuRowKey>("gpu_index");

  const handleRequestSort = (property: GpuRowKey) => {
    const isAsc = orderBy === property && order === "asc";
    setOrder(isAsc ? "desc" : "asc");
    setOrderBy(property);
  };

  const [rows, setRows] = React.useState<GPUMetrics[]>([]);
  React.useEffect(() => {
    fetchGpuMetricsWithFallback().then((res) => {
      setRows(res.data ?? []);
    });
  }, []);

  const sortedRows = React.useMemo(
    () => [...rows].sort(getComparator<GpuRow>(order, orderBy)),
    [rows, order, orderBy]
  );

  return (
    <TableContainer component={Paper}>
      <Table
        sx={{
          width: "100%",
          borderCollapse: "separate",
          borderSpacing: 0,
        }}
        aria-label="gpu table"
      >
        <TableHead>
          <TableRow>
            {columns.map((col, colIdx) => (
              <TableCell
                key={col.id}
                align={col.align || "left"}
                sortDirection={orderBy === col.id ? order : false}
                sx={{
                  borderRight:
                    colIdx !== columns.length - 1
                      ? "1px solid #e0e0e0"
                      : undefined,
                  // Optionally, add left border for first column
                  borderLeft: colIdx === 0 ? "1px solid #e0e0e0" : undefined,
                }}
              >
                <TableSortLabel
                  active={orderBy === col.id}
                  direction={orderBy === col.id ? order : "asc"}
                  onClick={() => handleRequestSort(col.id)}
                >
                  {col.label}
                </TableSortLabel>
              </TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>
          {sortedRows.map((row, idx) => (
            <TableRow key={row.node_name + "-" + row.gpu_index + "-" + idx}>
              {columns.map((col, colIdx) => (
                <TableCell
                  key={col.id}
                  align={col.align || "left"}
                  sx={{
                    borderRight:
                      colIdx !== columns.length - 1
                        ? "1px solid #e0e0e0"
                        : undefined,
                    borderLeft: colIdx === 0 ? "1px solid #e0e0e0" : undefined,
                  }}
                >
                  {row[col.id]}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
