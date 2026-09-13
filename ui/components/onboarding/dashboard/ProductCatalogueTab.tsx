import React, { useState } from "react";
import {
  ShoppingBag,
  Plus,
  Minus,
  Sparkles,
  AlertCircle,
  CheckCircle2,
  ShoppingCart,
  ArrowRight,
} from "lucide-react";
import { ProductItem, AppStatus } from "@/types/onboarding";
import { CreditOrderFlow } from "../CreditOrderFlow";
import { WholesaleCart, CartItem } from "./WholesaleCart";

interface ProductCatalogueTabProps {
  regularProducts: ProductItem[];
  appStatus?: AppStatus | null;
  token?: string;
  onOrderCreated?: () => void;
}

export const ProductCatalogueTab: React.FC<ProductCatalogueTabProps> = ({
  regularProducts,
  appStatus,
  token,
  onOrderCreated,
}) => {
  const catalogProducts = regularProducts.filter(
    (p) => !p.is_sample && p.is_regular !== false
  );

  // Quantity per product map (product.id -> quantity)
  const [quantities, setQuantities] = useState<Record<string, number>>({});
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // Active Order Credit Flow modal state
  const [activeCycleOrder, setActiveCycleOrder] = useState<{
    id: string;
    number: string;
    totalAmountPaise: number;
    advancePaidPaise: number;
  } | null>(null);

  const getQuantity = (product: ProductItem) => {
    return quantities[product.id] ?? (product.moq || 10);
  };

  const setQuantity = (productId: string, val: number, moq: number) => {
    const validVal = Math.max(moq || 1, val);
    setQuantities((prev) => ({ ...prev, [productId]: validVal }));
  };

  // Add product to cart
  const handleAddToCart = (product: ProductItem) => {
    const qty = getQuantity(product);
    setError(null);
    setSuccessMessage(`${product.name} (${qty} cases) added to cart.`);

    setCartItems((prev) => {
      const existingIdx = prev.findIndex((item) => item.product.id === product.id);
      if (existingIdx >= 0) {
        const updated = [...prev];
        updated[existingIdx].quantity += qty;
        return updated;
      }
      return [...prev, { product, quantity: qty }];
    });
  };

  const handleUpdateCartQty = (productId: string, qty: number) => {
    if (qty <= 0) {
      handleRemoveCartItem(productId);
      return;
    }
    setCartItems((prev) =>
      prev.map((item) =>
        item.product.id === productId ? { ...item, quantity: qty } : item
      )
    );
  };

  const handleRemoveCartItem = (productId: string) => {
    setCartItems((prev) => prev.filter((item) => item.product.id !== productId));
  };

  const handleClearCart = () => {
    setCartItems([]);
  };

  // Submit cart order to backend & launch CreditOrderFlow modal
  const handleCheckoutCart = async () => {
    if (cartItems.length === 0) return;

    const storedToken =
      token || localStorage.getItem("distributor_token") || localStorage.getItem("kresconet_token");

    if (!storedToken) {
      setError("Please sign in to place an order.");
      return;
    }

    setLoading(true);
    setError(null);
    setSuccessMessage(null);

    try {
      const payloadItems = cartItems.map((item) => ({
        product_id: item.product.id,
        quantity: item.quantity,
      }));

      const res = await fetch("http://localhost:8081/api/v1/orders", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${storedToken}`,
        },
        body: JSON.stringify({ items: payloadItems }),
      });

      const resBody = await res.json();
      if (!res.ok || resBody.success === false) {
        throw new Error(resBody.error?.message || resBody.message || "Failed to create order");
      }

      const orderData = resBody.data || resBody;
      if (!orderData || !orderData.id) {
        throw new Error("Invalid response: order ID is missing");
      }

      const cartTotalPaise = cartItems.reduce(
        (acc, item) => acc + item.product.price_paise * item.quantity,
        0
      );

      const totalVal =
        orderData.total_amount_paise && orderData.total_amount_paise > 0
          ? orderData.total_amount_paise
          : orderData.total_paise && orderData.total_paise > 0
          ? orderData.total_paise
          : cartTotalPaise;

      const advanceVal =
        orderData.advance_paid_paise !== undefined && orderData.advance_paid_paise !== null
          ? orderData.advance_paid_paise
          : orderData.advance_paise || 0;

      // Clear cart AFTER capturing order values
      setCartItems([]);

      // Launch Credit Cycle Flow Modal
      setActiveCycleOrder({
        id: orderData.id,
        number: orderData.order_number || `ORD-${orderData.id.slice(0, 6)}`,
        totalAmountPaise: totalVal,
        advancePaidPaise: advanceVal,
      });

      if (onOrderCreated) {
        onOrderCreated();
      }
    } catch (err: any) {
      setError(err.message || "An error occurred while placing cart order.");
    } finally {
      setLoading(false);
    }
  };

  const availableCreditPaise = appStatus?.assigned_credit_limit || 50000000;

  return (
    <div className="space-y-6 relative pb-28">
      {/* Banner / Feedback */}
      {error && (
        <div className="p-4 bg-rose-500/10 border border-rose-500/30 rounded-2xl flex items-center gap-3 text-xs text-rose-300">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {successMessage && (
        <div className="p-4 bg-emerald-500/10 border border-emerald-500/30 rounded-2xl flex items-center gap-3 text-xs text-emerald-300">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>{successMessage}</span>
        </div>
      )}

      <div className="bg-slate-900/70 border border-slate-800 rounded-2xl sm:rounded-3xl p-4 sm:p-6 shadow-xl space-y-5">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h3 className="text-base sm:text-lg font-bold text-white flex items-center gap-2">
              <ShoppingBag className="w-5 h-5 text-indigo-400" /> Product Catalogue & Wholesale Requirements
            </h3>
            <p className="text-xs text-slate-400 mt-1">
              Select products, add items to your cart, and preview weekly vs monthly repayment breakdowns.
            </p>
          </div>
          <span className="self-start sm:self-auto px-3 py-1 bg-indigo-500/10 text-indigo-300 border border-indigo-500/30 rounded-full text-xs font-bold flex items-center gap-1.5">
            <Sparkles className="w-3.5 h-3.5" /> Order Credit Cycle Enabled
          </span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3.5 sm:gap-4">
          {catalogProducts.map((p) => {
            const currentQty = getQuantity(p);
            const moq = p.moq || 10;
            const totalPricePaise = p.price_paise * currentQty;

            return (
              <div
                key={p.id}
                className="p-4 sm:p-5 bg-slate-950/80 border border-slate-800 rounded-xl sm:rounded-2xl space-y-4 hover:border-indigo-500/40 transition-all flex flex-col justify-between"
              >
                <div className="space-y-3">
                  {p.image_url ? (
                    <div className="w-full h-32 rounded-xl overflow-hidden bg-slate-900 border border-slate-800">
                      <img src={p.image_url} alt={p.name} className="w-full h-full object-cover" />
                    </div>
                  ) : (
                    <div className="w-full h-24 rounded-xl bg-slate-900/60 border border-slate-800 flex items-center justify-center text-indigo-400/40">
                      <ShoppingBag className="w-8 h-8" />
                    </div>
                  )}
                  <div className="flex justify-between items-start gap-2">
                    <div>
                      <h4 className="font-bold text-white text-sm leading-snug">{p.name}</h4>
                      <span className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider block mt-0.5">
                        {p.category}
                      </span>
                    </div>
                    <span className="text-[11px] font-bold text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full border border-emerald-500/20 shrink-0">
                      MOQ: {moq}
                    </span>
                  </div>
                </div>

                <div className="space-y-3 pt-3 border-t border-slate-800/80">
                  {/* Quantity Counter */}
                  <div className="flex items-center justify-between bg-slate-900/90 border border-slate-800 rounded-xl p-1.5">
                    <span className="text-[11px] text-slate-400 pl-2 font-medium">Quantity ({p.unit || "case"})</span>
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        onClick={() => setQuantity(p.id, currentQty - 1, moq)}
                        disabled={currentQty <= moq}
                        className="w-7 h-7 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-30 text-slate-200 flex items-center justify-center transition-colors"
                      >
                        <Minus className="w-3.5 h-3.5" />
                      </button>
                      <input
                        type="number"
                        min={moq}
                        value={currentQty}
                        onChange={(e) => setQuantity(p.id, parseInt(e.target.value) || moq, moq)}
                        className="w-12 text-center bg-transparent font-bold text-xs text-white focus:outline-none"
                      />
                      <button
                        type="button"
                        onClick={() => setQuantity(p.id, currentQty + 1, moq)}
                        className="w-7 h-7 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 flex items-center justify-center transition-colors"
                      >
                        <Plus className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>

                  {/* Price & Add to Cart Button */}
                  <div className="flex items-center justify-between gap-2 pt-1">
                    <div>
                      <div className="text-[10px] text-slate-500">Unit Total</div>
                      <div className="text-sm sm:text-base font-black text-emerald-400">
                        ₹{(totalPricePaise / 100).toLocaleString("en-IN")}
                      </div>
                    </div>

                    <button
                      type="button"
                      onClick={() => handleAddToCart(p)}
                      className="px-3.5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white text-xs font-bold shadow-md shadow-indigo-600/20 transition-all flex items-center gap-1.5 min-h-[38px] touch-manipulation"
                    >
                      <ShoppingCart className="w-3.5 h-3.5" />
                      <span>Add to Cart</span>
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Floating Wholesale Cart Drawer */}
      <WholesaleCart
        cartItems={cartItems}
        onUpdateQuantity={handleUpdateCartQty}
        onRemoveItem={handleRemoveCartItem}
        onClearCart={handleClearCart}
        onCheckout={handleCheckoutCart}
        loading={loading}
        availableCreditPaise={availableCreditPaise}
      />

      {/* Credit Order Flow Modal when order is submitted */}
      {activeCycleOrder && (
        <CreditOrderFlow
          orderId={activeCycleOrder.id}
          orderNumber={activeCycleOrder.number}
          totalAmountPaise={activeCycleOrder.totalAmountPaise}
          advancePaidPaise={activeCycleOrder.advancePaidPaise}
          availableCreditPaise={availableCreditPaise}
          token={token || localStorage.getItem("distributor_token") || localStorage.getItem("kresconet_token") || ""}
          onSuccess={() => {
            setActiveCycleOrder(null);
            setSuccessMessage("Order Credit Cycle active & mandate created!");
          }}
          onClose={() => setActiveCycleOrder(null)}
        />
      )}
    </div>
  );
};
