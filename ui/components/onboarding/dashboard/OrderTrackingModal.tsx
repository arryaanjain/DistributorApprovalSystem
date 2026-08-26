import React from "react";
import { CheckCircle2, Clock, Truck, X } from "lucide-react";

interface OrderTrackingModalProps {
  selectedOrder: any;
  selectedOrderType: "sample" | "commercial";
  onClose: () => void;
}

export const OrderTrackingModal: React.FC<OrderTrackingModalProps> = ({
  selectedOrder,
  selectedOrderType,
  onClose,
}) => {
  if (!selectedOrder) return null;

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/85 backdrop-blur-md flex items-center justify-center p-3 sm:p-4 overflow-y-auto">
      <div className="w-full max-w-xl bg-slate-900 border border-slate-800 rounded-2xl sm:rounded-3xl p-5 sm:p-6 shadow-2xl space-y-5 text-left my-auto max-h-[90vh] flex flex-col">
        {/* Modal Header */}
        <div className="flex items-center justify-between border-b border-slate-800 pb-3.5 shrink-0">
          <div>
            <span className="text-[10px] font-bold uppercase tracking-wider text-indigo-400">
              {selectedOrderType === "sample" ? "Sample Kit Trial" : "Commercial Order"}
            </span>
            <h3 className="text-base sm:text-lg font-bold text-white mt-0.5">
              Order Ref: #{selectedOrder.order_number || selectedOrder.id.slice(0, 8)}
            </h3>
          </div>

          <button
            onClick={onClose}
            type="button"
            className="p-2 rounded-xl bg-slate-800 text-slate-400 hover:text-white transition-all active:bg-slate-700 touch-manipulation min-w-[40px] min-h-[40px] flex items-center justify-center"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Modal Scrollable Body */}
        <div className="overflow-y-auto space-y-5 pr-1 flex-1">
          {/* Summary Banner */}
          <div className="bg-slate-950 p-4 rounded-xl sm:rounded-2xl border border-slate-800 grid grid-cols-2 gap-3 text-xs">
            <div>
              <span className="text-slate-400">Total Amount:</span>
              <div className="text-sm font-black text-white mt-0.5">
                ₹{((selectedOrderType === "sample" ? selectedOrder.amount_paise : selectedOrder.total_amount_paise) / 100).toLocaleString("en-IN")}
              </div>
            </div>
            <div>
              <span className="text-slate-400">Placed On:</span>
              <div className="text-xs font-semibold text-slate-200 mt-0.5">
                {new Date(selectedOrder.created_at).toLocaleDateString("en-IN")}
              </div>
            </div>
          </div>

          {/* Stepper Timeline */}
          <div className="space-y-4">
            <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400">
              Fulfillment Pipeline Timeline
            </h4>

            <div className="relative border-l-2 border-slate-800 pl-5 sm:pl-6 space-y-6 ml-2">
              {/* Step 1 */}
              <div className="relative">
                <div className="absolute -left-[27px] sm:-left-[31px] top-0.5 w-4 h-4 rounded-full bg-emerald-500 border-2 border-slate-900 flex items-center justify-center">
                  <CheckCircle2 className="w-3 h-3 text-slate-950" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">Order Submitted</div>
                  <div className="text-[11px] text-slate-400 mt-0.5">
                    {new Date(selectedOrder.created_at).toLocaleString("en-IN")}
                  </div>
                </div>
              </div>

              {/* Step 2 */}
              <div className="relative">
                <div className="absolute -left-[27px] sm:-left-[31px] top-0.5 w-4 h-4 rounded-full bg-emerald-500 border-2 border-slate-900 flex items-center justify-center">
                  <CheckCircle2 className="w-3 h-3 text-slate-950" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">
                    {selectedOrderType === "sample" ? "Razorpay Prepaid Verification" : "Credit Sanction Reserve"}
                  </div>
                  <div className="text-[11px] text-slate-400 mt-0.5">
                    {selectedOrder.razorpay_order_id
                      ? `RP Ref: ${selectedOrder.razorpay_order_id}`
                      : "Credit Balance Deducted"}
                  </div>
                </div>
              </div>

              {/* Step 3 */}
              <div className="relative">
                <div
                  className={`absolute -left-[27px] sm:-left-[31px] top-0.5 w-4 h-4 rounded-full border-2 border-slate-900 flex items-center justify-center ${
                    selectedOrder.status === "DISPATCHED" ||
                    selectedOrder.status === "APPROVED" ||
                    selectedOrder.status === "PAID"
                      ? "bg-emerald-500 text-slate-950"
                      : "bg-amber-500 text-slate-950"
                  }`}
                >
                  <Clock className="w-3 h-3" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">Admin Warehouse Approval</div>
                  <div className="text-[11px] text-slate-400 mt-0.5">
                    {selectedOrder.status === "PENDING_REVIEW"
                      ? "Awaiting Admin Sign-off"
                      : "Approved for Dispatch"}
                  </div>
                </div>
              </div>

              {/* Step 4 */}
              <div className="relative">
                <div
                  className={`absolute -left-[27px] sm:-left-[31px] top-0.5 w-4 h-4 rounded-full border-2 border-slate-900 flex items-center justify-center ${
                    selectedOrder.status === "DISPATCHED"
                      ? "bg-emerald-500 text-slate-950"
                      : "bg-slate-800 text-slate-500"
                  }`}
                >
                  <Truck className="w-3 h-3" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">Shiprocket Carrier Dispatch</div>
                  <div className="text-[11px] text-slate-400 mt-0.5">
                    {selectedOrder.status === "DISPATCHED"
                      ? "Courier AWB Assigned & Handed Over"
                      : "In Logistics Queue"}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Modal Footer */}
        <div className="pt-3 border-t border-slate-800 flex justify-end shrink-0">
          <button
            onClick={onClose}
            type="button"
            className="w-full sm:w-auto px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs rounded-xl transition-all active:bg-indigo-700 min-h-[44px] touch-manipulation flex items-center justify-center"
          >
            Close Tracking Detail
          </button>
        </div>
      </div>
    </div>
  );
};
