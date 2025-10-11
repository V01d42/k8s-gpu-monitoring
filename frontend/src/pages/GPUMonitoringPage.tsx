import GPUProcessesTable from "../components/GPUProcessesTable";
import GPUUsageTable from "../components/GPUUsageTable";

const GPUMonitoringPage = () => {
  return (
    <>
      <GPUUsageTable />
      <GPUProcessesTable />
    </>
  );
};

export default GPUMonitoringPage;
