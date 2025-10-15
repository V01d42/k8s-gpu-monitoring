import type { RouteObject } from "react-router-dom";
import GPUMonitoringPage from "./pages/GPUMonitoringPage";

const routes: RouteObject[] = [
  {
    children: [
      { path: "/", element: <GPUMonitoringPage /> },
      { path: "/healthz", element: <div>OK</div> },
    ],
  },
];

export default routes;
