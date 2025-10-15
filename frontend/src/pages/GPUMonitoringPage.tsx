import GPUProcessesTable from "../components/GPUProcessesTable";
import GPUUsageTable from "../components/GPUUsageTable";

const GPUMonitoringPage = () => {
  const components = [GPUUsageTable, GPUProcessesTable];
  return (
    <>
      {components.map((Comp) => (
        <Comp />
      ))}
    </>
  );
};

export default GPUMonitoringPage;
