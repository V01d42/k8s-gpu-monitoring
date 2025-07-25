import type { RouteObject } from 'react-router-dom';
import GPUTable from './components/GPUTable';


const routes: RouteObject[] = [
  {
     children: [
        { path: '/', element: <GPUTable /> },
        { path: '/healthz', element: <div>OK</div> },
    ],
  },
];

export default routes;