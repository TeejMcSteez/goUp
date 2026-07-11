import { useState, useMemo } from "react";
import useServiceData from "../hooks/useServiceData";
import ServiceCard from "./features/services/ServiceCard";
import ServiceControls from "./features/services/ServiceControls";

export default function Services() {
  const { data: services, loading, error } = useServiceData();
  const [searchTerm, setSearchTerm] = useState("");
  const [sortKey, setSortKey] = useState("name");

  const sortedAndFilteredServices = useMemo(() => {
    if (!services || services.length === 0) return [];

    const filtered = services.filter((service) =>
      service.name.toLowerCase().includes(searchTerm.toLowerCase()),
    );

    return [...filtered].sort((a, b) => {
      switch (sortKey) {
        case "name":
          return a.name.localeCompare(b.name);
        case "status":
          return a.error === b.error ? 0 : a.error ? -1 : 1;
        default:
          return 0;
      }
    });
  }, [services, searchTerm, sortKey]);

  const renderCards = () => {
    if (error) {
      return (
        <div className="bg-surface rounded-xl p-6 border border-border">
          <p>Error loading service data: {error}</p>
        </div>
      );
    }

    if (loading && !services) {
      return (
        <div className="bg-surface rounded-xl p-6 border border-border">
          <p>Loading services...</p>
        </div>
      );
    }

    if (!services || services.length === 0) {
      return (
        <div className="bg-surface rounded-xl p-6 border border-border">
          <p>No Service Data to Display</p>
        </div>
      );
    }

    if (sortedAndFilteredServices.length === 0) {
      return (
        <div className="bg-surface rounded-xl p-6 border border-border">
          <p>No services match your search.</p>
        </div>
      );
    }

    return sortedAndFilteredServices.map((svc) => (
      <ServiceCard key={svc.name} service={svc} />
    ));
  };

  return (
    <div className="w-full flex flex-col items-center text-center justify-center">
      <ServiceControls
        searchTerm={searchTerm}
        setSearchTerm={setSearchTerm}
        sortKey={sortKey}
        setSortKey={setSortKey}
      />
      <div
        id="cards"
        className="w-full py-4 pb-12 grid grid-cols-1 sm:grid-cols-[repeat(auto-fit,minmax(min(300px,100%),1fr))] gap-4 sm:gap-6"
      >
        {renderCards()}
      </div>
    </div>
  );
}
