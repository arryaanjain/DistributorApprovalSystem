import React, { useEffect, useState } from "react";
import { ProductItem, AppStatus } from "@/types/onboarding";
import { fetchApi } from "@/lib/api";

import { DashboardHeader } from "./dashboard/DashboardHeader";
import { DashboardNavTabs, DashboardTabType } from "./dashboard/DashboardNavTabs";
import { LiveOrderTrackingTab } from "./dashboard/LiveOrderTrackingTab";
import { ProductCatalogueTab } from "./dashboard/ProductCatalogueTab";
import { CreditFacilityTab } from "./dashboard/CreditFacilityTab";
import { AddressDirectoryTab } from "./dashboard/AddressDirectoryTab";
import { OrderTrackingModal } from "./dashboard/OrderTrackingModal";

interface Step9DashboardProps {
  trialActivated: boolean;
  regularProducts: ProductItem[];
  appStatus?: AppStatus | null;
  onSignOut?: () => void;
}

export const Step9Dashboard: React.FC<Step9DashboardProps> = ({
  trialActivated,
  regularProducts,
  appStatus,
  onSignOut,
}) => {
  const [activeTab, setActiveTab] = useState<DashboardTabType>("tracking");
  const [catalogOrders, setCatalogOrders] = useState<any[]>([]);
  const [sampleOrders, setSampleOrders] = useState<any[]>([]);
  const [loadingOrders, setLoadingOrders] = useState<boolean>(true);

  // Order Detail Modal State
  const [selectedOrder, setSelectedOrder] = useState<any | null>(null);
  const [selectedOrderType, setSelectedOrderType] = useState<"sample" | "commercial">("commercial");

  const loadMyOrders = async () => {
    setLoadingOrders(true);
    try {
      const [catRes, sampleRes] = await Promise.all([
        fetchApi<any[]>("/orders"),
        fetchApi<any[]>("/orders/samples/mine"),
      ]);
      if (catRes.success && Array.isArray(catRes.data)) {
        setCatalogOrders(catRes.data);
      }
      if (sampleRes.success && Array.isArray(sampleRes.data)) {
        setSampleOrders(sampleRes.data);
      }
    } catch (err) {
      console.error("Failed loading order history", err);
    } finally {
      setLoadingOrders(false);
    }
  };

  useEffect(() => {
    loadMyOrders();
  }, []);

  const totalOrdersCount = catalogOrders.length + sampleOrders.length;

  const handleSelectOrder = (order: any, type: "sample" | "commercial") => {
    setSelectedOrder(order);
    setSelectedOrderType(type);
  };

  return (
    <div className="space-y-6 sm:space-y-8 max-w-6xl mx-auto pb-12 px-2 sm:px-0">
      {/* 1. Header Banner */}
      <DashboardHeader
        trialActivated={trialActivated}
        appStatus={appStatus}
        totalOrdersCount={totalOrdersCount}
        onSignOut={onSignOut}
      />

      {/* 2. Navigation Tabs */}
      <DashboardNavTabs activeTab={activeTab} setActiveTab={setActiveTab} />

      {/* 3. Tab Contents */}
      {activeTab === "tracking" && (
        <LiveOrderTrackingTab
          sampleOrders={sampleOrders}
          catalogOrders={catalogOrders}
          loadingOrders={loadingOrders}
          onRefresh={loadMyOrders}
          onSelectOrder={handleSelectOrder}
        />
      )}

      {activeTab === "catalogue" && <ProductCatalogueTab regularProducts={regularProducts} />}

      {activeTab === "credit" && (
        <CreditFacilityTab trialActivated={trialActivated} appStatus={appStatus} />
      )}

      {activeTab === "address" && <AddressDirectoryTab />}

      {/* 4. Tracking Modal */}
      <OrderTrackingModal
        selectedOrder={selectedOrder}
        selectedOrderType={selectedOrderType}
        onClose={() => setSelectedOrder(null)}
      />
    </div>
  );
};
