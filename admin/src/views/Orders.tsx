import React, { useEffect, useState } from 'react';
import { ShoppingBag, RefreshCw, Truck, AlertCircle } from 'lucide-react';
import { api } from '../services/api';

export const Orders: React.FC = () => {
  const [orders, setOrders] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [actionMessage, setActionMessage] = useState<string>('');

  const loadOrders = async () => {
    setLoading(true);
    try {
      const data = await api.listOrders(50, 0);
      setOrders(data.orders || []);
    } catch (err) {
      console.error('Failed loading orders', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadOrders();
  }, []);

  const handleApprove = async (id: string) => {
    try {
      await api.approveOrder(id);
      setActionMessage(`Order #${id.slice(0, 8)} approved! Credit reserved.`);
      loadOrders();
    } catch (err: any) {
      setActionMessage(`Error approving order: ${err.message}`);
    }
  };

  const handleDispatch = async (id: string) => {
    try {
      await api.dispatchOrder(id);
      setActionMessage(`Order #${id.slice(0, 8)} marked as Dispatched! Invoice generated.`);
      loadOrders();
    } catch (err: any) {
      setActionMessage(`Error dispatching order: ${err.message}`);
    }
  };

  const formatINR = (paise?: number) => {
    if (!paise) return '₹0';
    return new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(paise / 100);
  };

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight">Catalogue Orders & Dispatch</h1>
          <p className="text-sm text-slate-400 mt-1">Review orders placed using distributor credit lines and manage warehouse dispatch.</p>
        </div>
        <button
          onClick={loadOrders}
          className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors border border-slate-700 flex items-center gap-2 text-xs font-semibold"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh Orders</span>
        </button>
      </div>

      {actionMessage && (
        <div className="p-3 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-300 text-xs flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0 text-indigo-400" />
          <span>{actionMessage}</span>
        </div>
      )}

      <div className="glass-panel p-6 rounded-2xl border border-slate-800">
        {loading ? (
          <div className="py-12 text-center text-slate-500 text-sm">Loading orders...</div>
        ) : orders.length === 0 ? (
          <div className="py-12 text-center text-slate-500 text-sm space-y-2">
            <ShoppingBag className="w-8 h-8 text-slate-600 mx-auto" />
            <p>No distributor catalogue orders yet.</p>
            <p className="text-xs text-slate-600">Once distributors sign agreements and place orders, they appear here for warehouse dispatch.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-300">
              <thead className="bg-slate-900/60 text-slate-400 text-xs uppercase tracking-wider border-b border-slate-800">
                <tr>
                  <th className="py-3.5 px-4 font-semibold">Order Ref</th>
                  <th className="py-3.5 px-4 font-semibold">Distributor</th>
                  <th className="py-3.5 px-4 font-semibold">Total Amount</th>
                  <th className="py-3.5 px-4 font-semibold">Credit Used</th>
                  <th className="py-3.5 px-4 font-semibold">Status</th>
                  <th className="py-3.5 px-4 font-semibold text-right">Dispatch Control</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {orders.map((o) => (
                  <tr key={o.id} className="hover:bg-slate-800/30 transition-colors">
                    <td className="py-3.5 px-4 font-mono font-bold text-indigo-300">#{o.order_number || o.id.slice(0, 8)}</td>
                    <td className="py-3.5 px-4 text-white font-bold">{o.distributor_name || 'Distributor'}</td>
                    <td className="py-3.5 px-4 font-bold text-white">{formatINR(o.total_amount_paise)}</td>
                    <td className="py-3.5 px-4 text-emerald-400 font-bold">{formatINR(o.credit_used_paise)}</td>
                    <td className="py-3.5 px-4">
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                        {o.status}
                      </span>
                    </td>
                    <td className="py-3.5 px-4 text-right space-x-2">
                      {o.status === 'PENDING_REVIEW' && (
                        <button
                          onClick={() => handleApprove(o.id)}
                          className="px-3 py-1.5 rounded-lg bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-300 text-xs font-semibold border border-emerald-500/30 transition-colors"
                        >
                          Approve Order
                        </button>
                      )}
                      {o.status === 'APPROVED' && (
                        <button
                          onClick={() => handleDispatch(o.id)}
                          className="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-lg shadow-indigo-600/20 transition-all flex items-center gap-1.5 inline-flex"
                        >
                          <Truck className="w-3.5 h-3.5" /> Dispatch Shipment
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
    </div>
  );
};
