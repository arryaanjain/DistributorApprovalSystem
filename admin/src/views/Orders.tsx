import React, { useEffect, useState } from 'react';
import {
  ShoppingBag,
  RefreshCw,
  Truck,
  AlertCircle,
  FileText,
  CheckCircle2,
  Package,
  Sparkles,
  ExternalLink,
  X,
  Wallet,
  ShieldAlert,
  Printer,
  Compass,
  Zap,
  User,
  MapPin,
} from 'lucide-react';
import { api } from '../services/api';

export const Orders: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'catalog' | 'sample'>('catalog');
  const [catalogOrders, setCatalogOrders] = useState<any[]>([]);
  const [sampleOrders, setSampleOrders] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [actionMessage, setActionMessage] = useState<string>('');

  // Commercial Order Receipt Modal
  const [selectedCatalogOrder, setSelectedCatalogOrder] = useState<any | null>(null);

  // Modal 1: Create Shipment (Package Dimensions & Weight)
  const [shipmentModalOrder, setShipmentModalOrder] = useState<any | null>(null);
  const [weight, setWeight] = useState<number | string>(0.5);
  const [length, setLength] = useState<number | string>(10);
  const [breadth, setBreadth] = useState<number | string>(10);
  const [height, setHeight] = useState<number | string>(10);
  const [paymentMethod, setPaymentMethod] = useState<string>('Prepaid');
  const [pickupLocation, setPickupLocation] = useState<string>('warehouse');
  const [creatingShipment, setCreatingShipment] = useState<boolean>(false);

  // Modal 2: Courier Selection & Dispatch (Wallet Check, AWB, Label, Manifest)
  const [courierModalOrder, setCourierModalOrder] = useState<any | null>(null);
  const [walletBalance, setWalletBalance] = useState<number | null>(null);
  const [couriers, setCouriers] = useState<any[]>([]);
  const [loadingCouriers, setLoadingCouriers] = useState<boolean>(false);
  const [selectedCourier, setSelectedCourier] = useState<any | null>(null);
  const [sortBy, setSortBy] = useState<'rate' | 'etd' | 'rating'>('rate');
  const [dispatching, setDispatching] = useState<boolean>(false);
  const [dispatchError, setDispatchError] = useState<string>('');

  // Post-Dispatch Action States
  const [awbCode, setAwbCode] = useState<string>('');
  const [courierName, setCourierName] = useState<string>('');
  const [, setLabelUrl] = useState<string>('');
  const [, setManifestUrl] = useState<string>('');
  const [trackingData, setTrackingData] = useState<any | null>(null);
  const [loadingTracking, setLoadingTracking] = useState<boolean>(false);
  const [generatingLabel, setGeneratingLabel] = useState<boolean>(false);
  const [generatingManifest, setGeneratingManifest] = useState<boolean>(false);

  const loadData = async () => {
    setLoading(true);
    try {
      if (activeTab === 'catalog') {
        const data = await api.listCatalogOrders(50, 0);
        setCatalogOrders(data.orders || []);
      } else {
        const data = await api.listSampleOrders(50, 0);
        setSampleOrders(data.sample_orders || []);
      }
    } catch (err) {
      console.error('Failed loading orders', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [activeTab]);

  const handleApproveCatalog = async (id: string) => {
    try {
      await api.approveOrder(id);
      setActionMessage(`Order #${id.slice(0, 8)} approved! Credit reserved.`);
      setSelectedCatalogOrder(null);
      loadData();
    } catch (err: any) {
      setActionMessage(`Error approving order: ${err.message}`);
    }
  };

  const handleDispatchCatalog = async (id: string) => {
    try {
      await api.dispatchOrder(id);
      setActionMessage(`Order #${id.slice(0, 8)} marked as Dispatched! Invoice generated.`);
      setSelectedCatalogOrder(null);
      loadData();
    } catch (err: any) {
      setActionMessage(`Error dispatching order: ${err.message}`);
    }
  };

  // Step 1: Open Shipment Modal
  const openShipmentModal = (order: any) => {
    setShipmentModalOrder(order);
    setWeight(order.package_weight || 0.5);
    setLength(order.package_length || 10);
    setBreadth(order.package_breadth || 10);
    setHeight(order.package_height || 10);
    setPaymentMethod('Prepaid');
    setPickupLocation('warehouse');
  };

  const handleCreateShipment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!shipmentModalOrder) return;
    setCreatingShipment(true);
    try {
      const finalWeight = effectiveBilledWeight > 0 ? effectiveBilledWeight : 0.5;
      const finalLength = numLength > 0 ? numLength : 10;
      const finalBreadth = numBreadth > 0 ? numBreadth : 10;
      const finalHeight = numHeight > 0 ? numHeight : 10;

      const res = await api.createShipment(shipmentModalOrder.id, {
        weight: finalWeight,
        length: finalLength,
        breadth: finalBreadth,
        height: finalHeight,
        payment_method: paymentMethod,
        pickup_location: pickupLocation || 'warehouse',
      });

      const updatedOrder = {
        ...shipmentModalOrder,
        shiprocket_order_id: res.shiprocket_order_id,
        shipment_id: res.shipment_id,
        package_weight: finalWeight,
        package_length: finalLength,
        package_breadth: finalBreadth,
        package_height: finalHeight,
      };

      setShipmentModalOrder(null);
      setActionMessage(`Shipment created in Shiprocket (Order ID: ${res.shiprocket_order_id}). Fetching live couriers...`);
      openCourierModal(updatedOrder);
      loadData();
    } catch (err: any) {
      setActionMessage(`Failed to create shipment: ${err.message}`);
    } finally {
      setCreatingShipment(false);
    }
  };

  // Step 2: Open Courier Selection & Dispatch Modal
  const openCourierModal = async (order: any) => {
    setCourierModalOrder(order);
    setSelectedCourier(null);
    setDispatchError('');
    setCouriers([]);
    setAwbCode(order.awb_code || '');
    setCourierName(order.courier_name || '');
    setLabelUrl(order.label_url || '');
    setManifestUrl(order.manifest_url || '');
    setTrackingData(null);

    // Fetch Wallet Balance & Available Couriers
    setLoadingCouriers(true);
    try {
      const [walletRes, couriersRes] = await Promise.allSettled([
        api.getWalletBalance(),
        api.getAvailableCouriers(order.id),
      ]);

      if (walletRes.status === 'fulfilled') {
        const balData = walletRes.value?.data || walletRes.value;
        const balAmount = balData?.data?.balance_amount || balData?.balance_amount || 5000;
        setWalletBalance(parseFloat(balAmount));
      }

      if (couriersRes.status === 'fulfilled') {
        const extractCouriers = (obj: any): any[] => {
          if (!obj) return [];
          if (Array.isArray(obj)) return obj;
          if (Array.isArray(obj.available_courier_companies)) return obj.available_courier_companies;
          if (obj.data) return extractCouriers(obj.data);
          return [];
        };

        const courierList = extractCouriers(couriersRes.value);
        setCouriers(courierList);
        if (courierList.length > 0) {
          setSelectedCourier(courierList[0]);
        }
      }
    } catch (err: any) {
      console.error('Error fetching couriers or wallet', err);
    } finally {
      setLoadingCouriers(false);
    }
  };

  // Sort couriers matrix
  const sortedCouriers = [...couriers].sort((a, b) => {
    if (sortBy === 'rate') {
      const rateA = a.rate || a.freight_charge || 0;
      const rateB = b.rate || b.freight_charge || 0;
      return rateA - rateB;
    }
    if (sortBy === 'etd') {
      const etdA = parseInt(a.estimated_delivery_days || a.etd || '999');
      const etdB = parseInt(b.estimated_delivery_days || b.etd || '999');
      return etdA - etdB;
    }
    if (sortBy === 'rating') {
      const ratA = a.rating || 0;
      const ratB = b.rating || 0;
      return ratB - ratA;
    }
    return 0;
  });

  // Step 2 Action: Assign Courier & Dispatch
  const handleAssignAndDispatch = async () => {
    if (!courierModalOrder || !selectedCourier) return;
    const courierRate = selectedCourier.rate || selectedCourier.freight_charge || 0;

    if (walletBalance !== null && walletBalance < courierRate) {
      setDispatchError(
        `Insufficient Shiprocket wallet balance! Required: ₹${courierRate.toFixed(
          2
        )}, Available: ₹${walletBalance.toFixed(2)}. Please recharge your wallet.`
      );
      return;
    }

    setDispatching(true);
    setDispatchError('');
    try {
      const courierId = selectedCourier.courier_company_id || selectedCourier.id;
      const res = await api.assignCourier(courierModalOrder.id, courierId, courierRate);

      setAwbCode(res.awb_code);
      setCourierName(res.courier_name);
      setActionMessage(
        `Order dispatched! AWB ${res.awb_code} generated for ${res.courier_name}.`
      );

      // Auto trigger pickup request
      await api.requestPickup(courierModalOrder.id).catch(() => {});

      loadData();
    } catch (err: any) {
      setDispatchError(err.message || 'Courier assignment failed');
    } finally {
      setDispatching(false);
    }
  };

  // Post-Dispatch Action: Generate Label PDF
  const handleDownloadLabel = async () => {
    if (!courierModalOrder) return;
    setGeneratingLabel(true);
    try {
      const res = await api.generateLabel(courierModalOrder.id);
      if (res.label_url) {
        setLabelUrl(res.label_url);
        window.open(res.label_url, '_blank');
      } else {
        setActionMessage('Label generation initiated.');
      }
    } catch (err: any) {
      setActionMessage(`Label generation error: ${err.message}`);
    } finally {
      setGeneratingLabel(false);
    }
  };

  // Post-Dispatch Action: Generate Manifest PDF
  const handleGenerateManifest = async () => {
    if (!courierModalOrder) return;
    setGeneratingManifest(true);
    try {
      const res = await api.generateManifest(courierModalOrder.id);
      if (res.manifest_url) {
        setManifestUrl(res.manifest_url);
        window.open(res.manifest_url, '_blank');
      } else {
        setActionMessage('Manifest generated successfully.');
      }
    } catch (err: any) {
      setActionMessage(`Manifest error: ${err.message}`);
    } finally {
      setGeneratingManifest(false);
    }
  };

  // Post-Dispatch Action: Track Shipment
  const handleTrackShipment = async () => {
    if (!courierModalOrder) return;
    setLoadingTracking(true);
    try {
      const res = await api.trackShipment(courierModalOrder.id);
      setTrackingData(res.tracking_data || res);
    } catch (err: any) {
      setActionMessage(`Tracking error: ${err.message}`);
    } finally {
      setLoadingTracking(false);
    }
  };

  const formatINR = (paise?: number) => {
    if (!paise) return '₹0';
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: 'INR',
      maximumFractionDigits: 0,
    }).format(paise / 100);
  };

  const numWeight = Math.max(0, parseFloat(String(weight)) || 0);
  const numLength = Math.max(0, parseFloat(String(length)) || 0);
  const numBreadth = Math.max(0, parseFloat(String(breadth)) || 0);
  const numHeight = Math.max(0, parseFloat(String(height)) || 0);

  const rawVolumetricWeight = (numLength * numBreadth * numHeight) / 5000;
  const volumetricWeight = rawVolumetricWeight > 0 ? rawVolumetricWeight.toFixed(3) : '0.000';
  const effectiveBilledWeight = Math.max(numWeight, rawVolumetricWeight);
  const billedWeight = effectiveBilledWeight > 0 ? effectiveBilledWeight.toFixed(3) : '0.000';
  const isVolumetricGoverned = rawVolumetricWeight > numWeight && rawVolumetricWeight > 0;

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight">
            Order Management & Logistics
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Segmented oversight for commercial credit catalogue orders and Shiprocket sample kit trial dispatches.
          </p>
        </div>
        <button
          onClick={loadData}
          className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors border border-slate-700 flex items-center gap-2 text-xs font-semibold"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh</span>
        </button>
      </div>

      {/* Segmented Toggle Bar */}
      <div className="flex items-center gap-2 p-1.5 bg-slate-900/90 border border-slate-800 rounded-2xl w-fit">
        <button
          onClick={() => setActiveTab('catalog')}
          className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold transition-all ${
            activeTab === 'catalog'
              ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/30'
              : 'text-slate-400 hover:text-slate-200'
          }`}
        >
          <ShoppingBag className="w-4 h-4" />
          <span>Commercial Catalog Orders</span>
        </button>
        <button
          onClick={() => setActiveTab('sample')}
          className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold transition-all ${
            activeTab === 'sample'
              ? 'bg-amber-500 text-slate-950 shadow-lg shadow-amber-500/30'
              : 'text-slate-400 hover:text-slate-200'
          }`}
        >
          <Sparkles className="w-4 h-4" />
          <span>Sample Kit Orders</span>
        </button>
      </div>

      {actionMessage && (
        <div className="p-3.5 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-300 text-xs flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0 text-indigo-400" />
            <span>{actionMessage}</span>
          </div>
          <button
            onClick={() => setActionMessage('')}
            className="text-indigo-400 hover:text-white text-xs"
          >
            Dismiss
          </button>
        </div>
      )}

      {/* TAB 1: COMMERCIAL CATALOG ORDERS */}
      {activeTab === 'catalog' && (
        <div className="glass-panel p-6 rounded-2xl border border-slate-800">
          {loading ? (
            <div className="py-12 text-center text-slate-500 text-sm">Loading catalog orders...</div>
          ) : catalogOrders.length === 0 ? (
            <div className="py-12 text-center text-slate-500 text-sm space-y-2">
              <ShoppingBag className="w-8 h-8 text-slate-600 mx-auto" />
              <p>No distributor commercial catalogue orders yet.</p>
              <p className="text-xs text-slate-600">
                Commercial orders placed with sanctioned credit lines appear here for state acknowledgment & receipt control.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm text-slate-300">
                <thead className="bg-slate-900/60 text-slate-400 text-xs uppercase tracking-wider border-b border-slate-800">
                  <tr>
                    <th className="py-3.5 px-4 font-semibold">Order Ref</th>
                    <th className="py-3.5 px-4 font-semibold">Distributor / Business</th>
                    <th className="py-3.5 px-4 font-semibold">Total Value</th>
                    <th className="py-3.5 px-4 font-semibold">Credit Used</th>
                    <th className="py-3.5 px-4 font-semibold">Advance Paid</th>
                    <th className="py-3.5 px-4 font-semibold">Status</th>
                    <th className="py-3.5 px-4 font-semibold text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60">
                  {catalogOrders.map((o) => (
                    <tr key={o.id} className="hover:bg-slate-800/30 transition-colors">
                      <td className="py-3.5 px-4 font-mono font-bold text-indigo-300">
                        #{o.order_number || o.id.slice(0, 8)}
                      </td>
                      <td className="py-3.5 px-4">
                        <div className="font-bold text-white">
                          {o.business_name || o.distributor_name || 'Distributor Partner'}
                        </div>
                        <div className="text-xs text-slate-400">{o.distributor_mobile}</div>
                      </td>
                      <td className="py-3.5 px-4 font-bold text-white">
                        {formatINR(o.total_amount_paise)}
                      </td>
                      <td className="py-3.5 px-4 text-emerald-400 font-bold">
                        {formatINR(o.credit_used_paise)}
                      </td>
                      <td className="py-3.5 px-4 text-amber-400 font-bold">
                        {formatINR(o.advance_paid_paise)}
                      </td>
                      <td className="py-3.5 px-4">
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                          {o.status}
                        </span>
                      </td>
                      <td className="py-3.5 px-4 text-right">
                        <button
                          onClick={() => setSelectedCatalogOrder(o)}
                          className="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-bold border border-slate-700 transition-all flex items-center gap-1.5 ml-auto"
                        >
                          <FileText className="w-3.5 h-3.5 text-indigo-400" />
                          <span>Receipt & Control</span>
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* TAB 2: SAMPLE KIT ORDERS */}
      {activeTab === 'sample' && (
        <div className="glass-panel p-6 rounded-2xl border border-slate-800 space-y-4">
          {loading ? (
            <div className="py-12 text-center text-slate-500 text-sm">Loading sample kit orders...</div>
          ) : sampleOrders.length === 0 ? (
            <div className="py-12 text-center text-slate-500 text-sm space-y-2">
              <Package className="w-8 h-8 text-amber-500/60 mx-auto" />
              <p>No sample kit trial orders found.</p>
              <p className="text-xs text-slate-600">
                Sample kit purchases completed via Razorpay appear here for Shiprocket dispatch.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm text-slate-300">
                <thead className="bg-slate-900/60 text-slate-400 text-xs uppercase tracking-wider border-b border-slate-800">
                  <tr>
                    <th className="py-3.5 px-4 font-semibold">Sample Order ID</th>
                    <th className="py-3.5 px-4 font-semibold">Distributor</th>
                    <th className="py-3.5 px-4 font-semibold">Razorpay Ref</th>
                    <th className="py-3.5 px-4 font-semibold">Shipping Address</th>
                    <th className="py-3.5 px-4 font-semibold">Status</th>
                    <th className="py-3.5 px-4 font-semibold text-right">Shiprocket Logistics</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60">
                  {sampleOrders.map((s) => (
                    <tr key={s.id} className="hover:bg-slate-800/30 transition-colors">
                      <td className="py-3.5 px-4 font-mono text-xs font-bold text-amber-400">
                        #{s.id.slice(0, 8)}
                      </td>
                      <td className="py-3.5 px-4">
                        <div className="font-bold text-white">{s.distributor_name || 'Distributor'}</div>
                        <div className="text-xs text-slate-400">{s.distributor_mobile}</div>
                      </td>
                      <td className="py-3.5 px-4 font-mono text-xs text-slate-400">{s.razorpay_order_id}</td>
                      <td className="py-3.5 px-4 text-xs text-slate-300 max-w-xs">
                        {s.shipping_address ? (
                          <div className="space-y-0.5">
                            <div className="font-medium text-slate-200">{s.shipping_address.address_line1}</div>
                            <div className="text-slate-400">
                              {s.shipping_address.city}, {s.shipping_address.state} - {s.shipping_address.pin}
                            </div>
                          </div>
                        ) : (
                          <span className="text-slate-500 italic">No specific shipping address</span>
                        )}
                      </td>
                      <td className="py-3.5 px-4">
                        <span
                          className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold ${
                            s.status === 'DISPATCHED'
                              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                              : s.status === 'PROCESSING'
                              ? 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20'
                              : s.status === 'PAID'
                              ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                              : 'bg-slate-800 text-slate-400'
                          }`}
                        >
                          {s.status}
                        </span>
                      </td>
                      <td className="py-3.5 px-4 text-right">
                        {s.awb_code ? (
                          <div className="flex items-center justify-end gap-2">
                            <button
                              onClick={() => openCourierModal(s)}
                              className="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-amber-300 text-xs font-bold border border-amber-500/30 transition-all flex items-center gap-1.5"
                            >
                              <Truck className="w-3.5 h-3.5 text-amber-400" />
                              <span>Logistics Hub</span>
                            </button>
                          </div>
                        ) : s.shiprocket_order_id ? (
                          <button
                            onClick={() => openCourierModal(s)}
                            className="px-3.5 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-600/20 transition-all inline-flex items-center gap-1.5"
                          >
                            <Truck className="w-3.5 h-3.5" />
                            <span>Assign Courier</span>
                          </button>
                        ) : (
                          <button
                            onClick={() => openShipmentModal(s)}
                            className="px-3.5 py-1.5 rounded-xl bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-400 hover:to-orange-400 text-slate-950 text-xs font-bold shadow-md shadow-amber-500/20 transition-all inline-flex items-center gap-1.5"
                          >
                            <Package className="w-3.5 h-3.5" />
                            <span>Create Shipment</span>
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* MODAL 1: CREATE SHIPMENT (PACKAGE WEIGHT, DIMENSIONS & BILLING DETAILS) */}
      {shipmentModalOrder && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-md flex items-center justify-center p-4">
          <div className="w-full max-w-4xl bg-slate-900 border border-amber-500/30 rounded-3xl p-6 shadow-2xl space-y-5 text-left max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <div className="flex items-center gap-2.5">
                <div className="p-2.5 rounded-2xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
                  <Package className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="text-lg font-bold text-white">Create Shiprocket Shipment</h3>
                  <p className="text-xs text-slate-400">Review billing & recipient details, configure package dimensions and dead weight</p>
                </div>
              </div>
              <button
                onClick={() => setShipmentModalOrder(null)}
                className="p-1.5 rounded-xl bg-slate-800 text-slate-400 hover:text-white"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-12 gap-5 text-xs">
              {/* LEFT COLUMN: Billing & Product Details */}
              <div className="md:col-span-5 space-y-4">
                {/* Billing & Recipient Info Card */}
                <div className="bg-slate-950 p-4 rounded-2xl border border-slate-800 space-y-3">
                  <div className="flex items-center justify-between border-b border-slate-800/80 pb-2">
                    <span className="font-semibold text-slate-200 flex items-center gap-1.5">
                      <User className="w-4 h-4 text-amber-400" />
                      Billing & Recipient Info
                    </span>
                    <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20">
                      {shipmentModalOrder.status || 'PAID'}
                    </span>
                  </div>

                  <div className="space-y-2">
                    <div>
                      <div className="text-[11px] text-slate-400">Recipient Name</div>
                      <div className="font-bold text-white text-sm">
                        {shipmentModalOrder.distributor_name || 'Distributor Partner'}
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-2 text-[11px]">
                      <div>
                        <div className="text-slate-400">Phone</div>
                        <div className="font-medium text-slate-200">
                          {shipmentModalOrder.shipping_address?.phone || shipmentModalOrder.distributor_mobile || 'N/A'}
                        </div>
                      </div>
                      <div>
                        <div className="text-slate-400">Email</div>
                        <div className="font-medium text-slate-200 truncate" title={shipmentModalOrder.distributor_email || ''}>
                          {shipmentModalOrder.distributor_email || 'N/A'}
                        </div>
                      </div>
                    </div>

                    <div className="pt-1">
                      <div className="text-[11px] text-slate-400 mb-1 flex items-center gap-1">
                        <MapPin className="w-3.5 h-3.5 text-amber-400 shrink-0" />
                        <span>Delivery Address</span>
                      </div>
                      <div className="p-2.5 rounded-xl bg-slate-900 border border-slate-800/80 space-y-0.5">
                        <div className="font-medium text-slate-200">
                          {shipmentModalOrder.shipping_address?.address_line1 || 'Default Warehouse Delivery'}
                        </div>
                        {shipmentModalOrder.shipping_address?.address_line2 && (
                          <div className="text-slate-400">{shipmentModalOrder.shipping_address.address_line2}</div>
                        )}
                        <div className="text-amber-400/90 font-medium pt-0.5">
                          {shipmentModalOrder.shipping_address ? (
                            `${shipmentModalOrder.shipping_address.city}, ${shipmentModalOrder.shipping_address.state} - ${shipmentModalOrder.shipping_address.pin}`
                          ) : (
                            'New Delhi, Delhi - 110001'
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Product / Sample Order Summary Card */}
                <div className="bg-slate-950 p-4 rounded-2xl border border-slate-800 space-y-3">
                  <div className="flex items-center justify-between border-b border-slate-800/80 pb-2">
                    <span className="font-semibold text-slate-200 flex items-center gap-1.5">
                      <ShoppingBag className="w-4 h-4 text-indigo-400" />
                      Product Summary
                    </span>
                    <span className="text-slate-400 text-[10px] font-mono">
                      {formatINR(shipmentModalOrder.amount_paise || 50000)}
                    </span>
                  </div>

                  <div className="p-3 rounded-xl bg-slate-900 border border-slate-800 flex items-center justify-between">
                    <div className="space-y-0.5">
                      <div className="font-bold text-white">Distributor Sample Kit</div>
                      <div className="text-[10px] text-slate-400 font-mono">SKU: SAMPLE-KIT-01 | Qty: 1</div>
                    </div>
                    <div className="text-right">
                      <div className="font-bold text-emerald-400">
                        {formatINR(shipmentModalOrder.amount_paise || 50000)}
                      </div>
                      <div className="text-[10px] text-slate-400">Prepaid Order</div>
                    </div>
                  </div>

                  {shipmentModalOrder.razorpay_payment_id && (
                    <div className="text-[11px] text-slate-400 flex items-center justify-between px-1">
                      <span>Payment Ref:</span>
                      <span className="font-mono text-slate-300">{shipmentModalOrder.razorpay_payment_id}</span>
                    </div>
                  )}
                </div>
              </div>

              {/* RIGHT COLUMN: Package Parameters & Pickup Form */}
              <div className="md:col-span-7">
                <form onSubmit={handleCreateShipment} className="space-y-4">
                  <div className="bg-slate-950 p-4 rounded-2xl border border-slate-800 space-y-3">
                    <div className="font-semibold text-slate-300 flex items-center justify-between border-b border-slate-800/80 pb-2">
                      <span>Package Parameters & Pickup</span>
                      <span className="text-[10px] text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded-md border border-amber-500/20 font-mono">
                        Shiprocket API
                      </span>
                    </div>

                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="text-slate-400 block mb-1 font-medium">Dead Weight (kg)</label>
                        <input
                          type="number"
                          step="0.01"
                          min="0.01"
                          value={weight}
                          onChange={(e) => setWeight(e.target.value)}
                          className="w-full bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-white font-bold focus:border-amber-500 outline-none"
                          required
                        />
                      </div>

                      <div>
                        <label className="text-slate-400 block mb-1 font-medium">Volumetric Weight</label>
                        <div className="bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-amber-400 font-bold flex items-center justify-between">
                          <span>~{volumetricWeight} kg</span>
                          <span className="text-[10px] font-mono text-slate-500">(L×B×H/5000)</span>
                        </div>
                      </div>
                    </div>

                    {/* Chargeable Billed Weight Card */}
                    <div className="p-3 bg-slate-900/90 border border-slate-800 rounded-xl flex items-center justify-between">
                      <div>
                        <span className="text-slate-400 block text-[11px] font-medium">Billed Chargeable Weight</span>
                        <span className="text-base font-black text-emerald-400">{billedWeight} kg</span>
                      </div>
                      {isVolumetricGoverned ? (
                        <span className="px-2.5 py-1 rounded-full text-[10px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20">
                          ⚡ Volumetric Governed
                        </span>
                      ) : (
                        <span className="px-2.5 py-1 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                          ✓ Dead Weight Governed
                        </span>
                      )}
                    </div>

                    <div className="grid grid-cols-3 gap-2 pt-1">
                      <div>
                        <label className="text-slate-400 block mb-1">Length (cm)</label>
                        <input
                          type="number"
                          step="0.1"
                          min="0.1"
                          value={length}
                          onChange={(e) => setLength(e.target.value)}
                          className="w-full bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-white outline-none focus:border-amber-500 font-medium"
                          required
                        />
                      </div>
                      <div>
                        <label className="text-slate-400 block mb-1">Breadth (cm)</label>
                        <input
                          type="number"
                          step="0.1"
                          min="0.1"
                          value={breadth}
                          onChange={(e) => setBreadth(e.target.value)}
                          className="w-full bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-white outline-none focus:border-amber-500 font-medium"
                          required
                        />
                      </div>
                      <div>
                        <label className="text-slate-400 block mb-1">Height (cm)</label>
                        <input
                          type="number"
                          step="0.1"
                          min="0.1"
                          value={height}
                          onChange={(e) => setHeight(e.target.value)}
                          className="w-full bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-white outline-none focus:border-amber-500 font-medium"
                          required
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-3 pt-2">
                      <div>
                        <label className="text-slate-400 block mb-1 font-medium">Payment Mode</label>
                        <select
                          value={paymentMethod}
                          onChange={(e) => setPaymentMethod(e.target.value)}
                          className="w-full bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-white outline-none"
                        >
                          <option value="Prepaid">Prepaid (Sample Kit)</option>
                          <option value="COD">Cash on Delivery (COD)</option>
                        </select>
                      </div>

                      <div>
                        <label className="text-slate-400 block mb-1 font-medium">Pickup Warehouse</label>
                        <input
                          type="text"
                          value={pickupLocation}
                          onChange={(e) => setPickupLocation(e.target.value)}
                          className="w-full bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-white outline-none"
                          placeholder="warehouse"
                        />
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center justify-end gap-3 pt-2 border-t border-slate-800">
                    <button
                      type="button"
                      onClick={() => setShipmentModalOrder(null)}
                      className="px-4 py-2 rounded-xl bg-slate-800 text-slate-300 hover:text-white font-semibold"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={creatingShipment}
                      className="px-5 py-2.5 rounded-xl bg-amber-500 hover:bg-amber-400 text-slate-950 font-bold flex items-center gap-2 shadow-lg shadow-amber-500/20 disabled:opacity-50"
                    >
                      <Truck className="w-4 h-4" />
                      <span>{creatingShipment ? 'Creating Shipment...' : 'Create Shipment'}</span>
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* MODAL 2: COURIER SELECTION & DISPATCH (WALLET CHECK, AWB, LABELS) */}
      {courierModalOrder && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-md flex items-center justify-center p-4">
          <div className="w-full max-w-4xl bg-slate-900 border border-slate-800 rounded-3xl p-6 shadow-2xl space-y-5 text-left max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <div className="flex items-center gap-3">
                <div className="p-2.5 rounded-2xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                  <Truck className="w-6 h-6" />
                </div>
                <div>
                  <h3 className="text-lg font-bold text-white">Logistics Hub & Courier Dispatch</h3>
                  <p className="text-xs text-slate-400">
                    Order Ref: #{courierModalOrder.id.slice(0, 8)} | Shipment: #{courierModalOrder.shipment_id || 'Pending'}
                  </p>
                </div>
              </div>
              <button
                onClick={() => setCourierModalOrder(null)}
                className="p-1.5 rounded-xl bg-slate-800 text-slate-400 hover:text-white"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Shiprocket Wallet Card */}
            <div className="bg-slate-950 p-4 rounded-2xl border border-slate-800 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  <Wallet className="w-5 h-5" />
                </div>
                <div>
                  <div className="text-xs text-slate-400">Shiprocket Live Wallet Balance</div>
                  <div className="text-lg font-black text-white">
                    {walletBalance !== null ? `₹${walletBalance.toFixed(2)}` : 'Fetching...'}
                  </div>
                </div>
              </div>
              <div className="text-xs text-slate-400 text-right space-y-0.5">
                <div>
                  Weight: <span className="text-slate-200 font-semibold">{courierModalOrder.package_weight || 0.5} kg</span>
                  {' | '}
                  Volumetric: <span className="text-amber-400 font-semibold">{(((courierModalOrder.package_length || 10) * (courierModalOrder.package_breadth || 10) * (courierModalOrder.package_height || 10)) / 5000).toFixed(3)} kg</span>
                </div>
                <div>Dim: {courierModalOrder.package_length || 10}x{courierModalOrder.package_breadth || 10}x{courierModalOrder.package_height || 10} cm</div>
              </div>
            </div>

            {dispatchError && (
              <div className="p-3.5 rounded-xl bg-red-500/10 border border-red-500/30 text-red-300 text-xs flex items-center gap-2">
                <ShieldAlert className="w-5 h-5 shrink-0 text-red-400" />
                <span>{dispatchError}</span>
              </div>
            )}

            {/* Post-Dispatch State Card */}
            {awbCode ? (
              <div className="bg-indigo-500/10 border border-indigo-500/30 p-5 rounded-2xl space-y-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <CheckCircle2 className="w-5 h-5 text-emerald-400" />
                    <div>
                      <div className="font-bold text-white text-sm">Dispatched via {courierName || 'Shiprocket Partner'}</div>
                      <div className="text-xs text-indigo-300 font-mono">AWB Tracking Code: {awbCode}</div>
                    </div>
                  </div>
                  <span className="px-3 py-1 rounded-full text-xs font-bold bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                    DISPATCHED
                  </span>
                </div>

                <div className="grid grid-cols-3 gap-3 pt-2">
                  <button
                    onClick={handleDownloadLabel}
                    disabled={generatingLabel}
                    className="p-3 rounded-xl bg-slate-800 hover:bg-slate-700 text-white font-bold text-xs border border-slate-700 flex items-center justify-center gap-2 transition-all"
                  >
                    <Printer className="w-4 h-4 text-indigo-400" />
                    <span>{generatingLabel ? 'Generating...' : 'Download Label PDF'}</span>
                  </button>

                  <button
                    onClick={handleGenerateManifest}
                    disabled={generatingManifest}
                    className="p-3 rounded-xl bg-slate-800 hover:bg-slate-700 text-white font-bold text-xs border border-slate-700 flex items-center justify-center gap-2 transition-all"
                  >
                    <FileText className="w-4 h-4 text-amber-400" />
                    <span>{generatingManifest ? 'Generating...' : 'Generate Manifest PDF'}</span>
                  </button>

                  <button
                    onClick={handleTrackShipment}
                    disabled={loadingTracking}
                    className="p-3 rounded-xl bg-slate-800 hover:bg-slate-700 text-white font-bold text-xs border border-slate-700 flex items-center justify-center gap-2 transition-all"
                  >
                    <Compass className={`w-4 h-4 text-emerald-400 ${loadingTracking ? 'animate-spin' : ''}`} />
                    <span>{loadingTracking ? 'Tracking...' : 'Track Live Status'}</span>
                  </button>
                </div>

                {trackingData && (
                  <div className="mt-3 p-4 bg-slate-950 rounded-xl border border-slate-800 text-xs space-y-2">
                    <div className="font-bold text-white flex items-center gap-2">
                      <Zap className="w-4 h-4 text-amber-400" /> Live Shipment Timeline
                    </div>
                    {trackingData.shipment_track_activities && trackingData.shipment_track_activities.length > 0 ? (
                      <div className="space-y-1.5 pt-2">
                        {trackingData.shipment_track_activities.map((act: any, idx: number) => (
                          <div key={idx} className="flex items-center justify-between border-b border-slate-800/60 pb-1 text-slate-300">
                            <div>
                              <span className="font-bold text-indigo-400">{act.status}</span>: {act.activity} ({act.location})
                            </div>
                            <div className="text-[11px] text-slate-500 font-mono">{act.date}</div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="text-slate-400 text-xs italic">Package handed over to courier. Status updates will appear here.</div>
                    )}
                  </div>
                )}
              </div>
            ) : (
              /* Courier Serviceability Matrix */
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="font-bold text-white text-sm">Available Courier Partners</h4>
                  <div className="flex items-center gap-2 text-xs">
                    <span className="text-slate-400">Sort by:</span>
                    <button
                      onClick={() => setSortBy('rate')}
                      className={`px-2.5 py-1 rounded-lg text-xs font-semibold ${
                        sortBy === 'rate' ? 'bg-indigo-600 text-white' : 'bg-slate-800 text-slate-400'
                      }`}
                    >
                      Cheapest
                    </button>
                    <button
                      onClick={() => setSortBy('etd')}
                      className={`px-2.5 py-1 rounded-lg text-xs font-semibold ${
                        sortBy === 'etd' ? 'bg-indigo-600 text-white' : 'bg-slate-800 text-slate-400'
                      }`}
                    >
                      Fastest
                    </button>
                    <button
                      onClick={() => setSortBy('rating')}
                      className={`px-2.5 py-1 rounded-lg text-xs font-semibold ${
                        sortBy === 'rating' ? 'bg-indigo-600 text-white' : 'bg-slate-800 text-slate-400'
                      }`}
                    >
                      Top Rated
                    </button>
                  </div>
                </div>

                {loadingCouriers ? (
                  <div className="py-12 text-center text-slate-500 text-xs">Fetching live courier rates from Shiprocket...</div>
                ) : couriers.length === 0 ? (
                  <div className="py-8 text-center text-slate-500 text-xs">
                    No available couriers returned for this pin code. Please verify address details.
                  </div>
                ) : (
                  <div className="overflow-x-auto border border-slate-800 rounded-2xl max-h-72 overflow-y-auto">
                    <table className="w-full text-left text-xs text-slate-300">
                      <thead className="bg-slate-950 text-slate-400 uppercase tracking-wider font-semibold sticky top-0 border-b border-slate-800 text-[11px]">
                        <tr>
                          <th className="py-3 px-3">Courier Partner</th>
                          <th className="py-3 px-3">Rating (Radar)</th>
                          <th className="py-3 px-3">Expected Pickup</th>
                          <th className="py-3 px-3">Estimated Delivery</th>
                          <th className="py-3 px-3">Chargeable Wt</th>
                          <th className="py-3 px-3">Charges</th>
                          <th className="py-3 px-3 text-right">Action</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-800/60 bg-slate-900/60">
                        {sortedCouriers.map((c, idx) => {
                          const isSelected =
                            selectedCourier?.courier_company_id === c.courier_company_id ||
                            selectedCourier?.id === c.id;
                          const rate = c.rate || c.freight_charge || 0;
                          const etd = c.etd || (c.estimated_delivery_days ? `${c.estimated_delivery_days} Days` : 'Aug 28, 2026');
                          const minWeight = c.min_weight || 0.5;
                          const rto = c.rto_charges || 100;
                          const rating = c.rating || 4.5;
                          const pickup = c.expected_pickup || 'Today';
                          const isRecommended = c.is_recommended || idx === 0;
                          const chargeableWeight = c.min_weight ? `${c.min_weight} Kg` : (courierModalOrder ? `${Math.max(courierModalOrder.package_weight || 0.5, 0.5).toFixed(1)} Kg` : '1 Kg');

                          return (
                            <tr
                              key={c.courier_company_id || c.id || idx}
                              onClick={() => setSelectedCourier(c)}
                              className={`cursor-pointer transition-colors border-b border-slate-800/40 ${
                                isSelected ? 'bg-indigo-600/20 text-white' : 'hover:bg-slate-800/40'
                              }`}
                            >
                              <td className="py-3 px-3">
                                <div className="flex items-center gap-2">
                                  {isRecommended && (
                                    <span className="px-2 py-0.5 rounded text-[10px] font-extrabold bg-gradient-to-r from-amber-500 to-orange-500 text-slate-950 shadow-sm shrink-0">
                                      Recommended
                                    </span>
                                  )}
                                  <div>
                                    <div className="font-bold text-white text-xs">{c.courier_name}</div>
                                    <div className="text-[11px] text-slate-400 flex items-center gap-1.5 pt-0.5">
                                      <span>{c.is_surface ? 'Surface' : 'Air'} | Min-weight: {minWeight} Kg</span>
                                      <span className="text-slate-600">|</span>
                                      <span className="text-rose-400/90 font-mono">RTO Charges: ₹{Number(rto).toFixed(2)}</span>
                                    </div>
                                  </div>
                                </div>
                              </td>
                              <td className="py-3 px-3">
                                <div className="flex items-center gap-1 font-bold text-amber-400 text-xs">
                                  <span>★ {rating}</span>
                                </div>
                              </td>
                              <td className="py-3 px-3 text-xs text-slate-300 font-medium">{pickup}</td>
                              <td className="py-3 px-3 text-xs text-emerald-400 font-semibold">{etd}</td>
                              <td className="py-3 px-3 text-xs font-mono text-slate-300">{chargeableWeight}</td>
                              <td className="py-3 px-3 font-black text-white text-xs">
                                ₹{Number(rate).toFixed(2)}
                              </td>
                              <td className="py-3 px-3 text-right">
                                <input
                                  type="radio"
                                  checked={isSelected}
                                  onChange={() => setSelectedCourier(c)}
                                  className="accent-indigo-500 w-4 h-4 cursor-pointer"
                                />
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                )}

                <div className="flex items-center justify-between pt-3 border-t border-slate-800">
                  <div className="text-xs text-slate-400">
                    Selected Courier Rate:{' '}
                    <span className="font-bold text-white">
                      ₹{(selectedCourier?.rate || selectedCourier?.freight_charge || 0).toFixed(2)}
                    </span>
                  </div>
                  <button
                    onClick={handleAssignAndDispatch}
                    disabled={dispatching || !selectedCourier}
                    className="px-6 py-2.5 rounded-xl bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-400 hover:to-orange-400 text-slate-950 font-bold text-xs flex items-center gap-2 shadow-lg shadow-amber-500/20 disabled:opacity-50"
                  >
                    <Truck className="w-4 h-4" />
                    <span>{dispatching ? 'Assigning AWB & Pickup...' : 'Assign Courier & Dispatch'}</span>
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* COMMERCIAL RECEIPT MODAL */}
      {selectedCatalogOrder && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-md flex items-center justify-center p-4">
          <div className="w-full max-w-lg bg-slate-900 border border-slate-800 rounded-3xl p-6 shadow-2xl space-y-5 text-left">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <div>
                <h3 className="text-lg font-bold text-white">Commercial Order Receipt</h3>
                <p className="text-xs text-indigo-400 font-mono">
                  #{selectedCatalogOrder.order_number || selectedCatalogOrder.id}
                </p>
              </div>
              <button
                onClick={() => setSelectedCatalogOrder(null)}
                className="p-1.5 rounded-xl bg-slate-800 text-slate-400 hover:text-white"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-3 text-xs">
              <div className="bg-slate-950 p-4 rounded-2xl border border-slate-800 space-y-2">
                <div className="flex justify-between">
                  <span className="text-slate-400">Distributor Business:</span>
                  <span className="font-bold text-white">
                    {selectedCatalogOrder.business_name || selectedCatalogOrder.distributor_name || 'N/A'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Total Order Amount:</span>
                  <span className="font-bold text-white">{formatINR(selectedCatalogOrder.total_amount_paise)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Credit Amount Used:</span>
                  <span className="font-bold text-emerald-400">
                    {formatINR(selectedCatalogOrder.credit_used_paise)}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Advance Paid:</span>
                  <span className="font-bold text-amber-400">
                    {formatINR(selectedCatalogOrder.advance_paid_paise)}
                  </span>
                </div>
                {selectedCatalogOrder.utr_reference && (
                  <div className="flex justify-between">
                    <span className="text-slate-400">UTR Reference:</span>
                    <span className="font-mono text-indigo-300">{selectedCatalogOrder.utr_reference}</span>
                  </div>
                )}
              </div>

              {selectedCatalogOrder.payment_proof_url && (
                <div className="p-3 bg-slate-800/40 rounded-xl border border-slate-700/60 flex items-center justify-between">
                  <span className="text-slate-300 font-medium">Payment Proof Attachment</span>
                  <a
                    href={selectedCatalogOrder.payment_proof_url}
                    target="_blank"
                    rel="noreferrer"
                    className="text-indigo-400 hover:underline flex items-center gap-1 font-bold"
                  >
                    View Document <ExternalLink className="w-3.5 h-3.5" />
                  </a>
                </div>
              )}
            </div>

            <div className="pt-2 border-t border-slate-800 flex items-center justify-end gap-3">
              {selectedCatalogOrder.status === 'PENDING_REVIEW' && (
                <button
                  onClick={() => handleApproveCatalog(selectedCatalogOrder.id)}
                  className="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-xs transition-all shadow-lg shadow-emerald-600/20"
                >
                  Approve Commercial Order
                </button>
              )}

              {selectedCatalogOrder.status === 'APPROVED' && (
                <button
                  onClick={() => handleDispatchCatalog(selectedCatalogOrder.id)}
                  className="px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs transition-all shadow-lg shadow-indigo-600/20 flex items-center gap-1.5"
                >
                  <Truck className="w-4 h-4" /> Dispatch Warehouse Shipment
                </button>
              )}

              <button
                onClick={() => setSelectedCatalogOrder(null)}
                className="px-4 py-2 rounded-xl bg-slate-800 text-slate-300 hover:text-white font-semibold text-xs"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
