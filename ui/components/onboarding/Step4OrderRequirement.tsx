import React, { useState, useEffect } from "react";
import {
  ShoppingCart,
  CheckCircle2,
  ChevronRight,
  ChevronDown,
  ChevronUp,
  Sparkles,
  Package,
  ArrowRight,
  Image as ImageIcon,
  Info,
  ZoomIn,
  ZoomOut,
  RotateCcw,
  X,
  Maximize2,
} from "lucide-react";
import { ProductItem } from "@/types/onboarding";

interface Step4OrderRequirementProps {
  orderChoice: "none" | "full" | "sample";
  setOrderChoice: (val: "none" | "full" | "sample") => void;
  regularProducts: ProductItem[];
  sampleProducts: ProductItem[];
  orderQuantities: Record<string, number>;
  handleQuantityChange: (id: string, delta: number) => void;
  calculateOrderTotal: () => number;
  onOpenOrderReview: () => void;
  onInitiateSampleBooking: (item: ProductItem) => void;
}

interface ImageZoomModalProps {
  url: string;
  title: string;
  onClose: () => void;
}

const ImageZoomModal: React.FC<ImageZoomModalProps> = ({ url, title, onClose }) => {
  const [scale, setScale] = useState(1);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [touchDist, setTouchDist] = useState<number | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const handleZoomIn = () => setScale((s) => Math.min(s + 0.5, 4));
  const handleZoomOut = () => {
    setScale((s) => {
      const next = Math.max(s - 0.5, 1);
      if (next === 1) setPosition({ x: 0, y: 0 });
      return next;
    });
  };

  const handleReset = () => {
    setScale(1);
    setPosition({ x: 0, y: 0 });
  };

  const handleWheel = (e: React.WheelEvent) => {
    e.preventDefault();
    if (e.deltaY < 0) {
      setScale((s) => Math.min(s + 0.25, 4));
    } else {
      setScale((s) => {
        const next = Math.max(s - 0.25, 1);
        if (next === 1) setPosition({ x: 0, y: 0 });
        return next;
      });
    }
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    if (scale > 1) {
      setIsDragging(true);
      setDragStart({ x: e.clientX - position.x, y: e.clientY - position.y });
    }
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (isDragging && scale > 1) {
      setPosition({
        x: e.clientX - dragStart.x,
        y: e.clientY - dragStart.y,
      });
    }
  };

  const handleMouseUp = () => setIsDragging(false);

  const handleTouchStart = (e: React.TouchEvent) => {
    if (e.touches.length === 2) {
      const dist = Math.hypot(
        e.touches[0].clientX - e.touches[1].clientX,
        e.touches[0].clientY - e.touches[1].clientY
      );
      setTouchDist(dist);
    } else if (e.touches.length === 1 && scale > 1) {
      setIsDragging(true);
      setDragStart({
        x: e.touches[0].clientX - position.x,
        y: e.touches[0].clientY - position.y,
      });
    }
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    if (e.touches.length === 2 && touchDist !== null) {
      const newDist = Math.hypot(
        e.touches[0].clientX - e.touches[1].clientX,
        e.touches[0].clientY - e.touches[1].clientY
      );
      const factor = newDist / touchDist;
      setScale((s) => {
        const next = Math.min(Math.max(s * factor, 1), 4);
        if (next === 1) setPosition({ x: 0, y: 0 });
        return next;
      });
      setTouchDist(newDist);
    } else if (e.touches.length === 1 && isDragging && scale > 1) {
      setPosition({
        x: e.touches[0].clientX - dragStart.x,
        y: e.touches[0].clientY - dragStart.y,
      });
    }
  };

  const handleTouchEnd = () => {
    setTouchDist(null);
    setIsDragging(false);
  };

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/95 backdrop-blur-md flex flex-col justify-between p-4 sm:p-6 select-none animate-fadeIn">
      {/* Header Controls */}
      <div className="flex items-center justify-between z-20 bg-slate-900/90 border border-slate-800 px-4 py-3 rounded-2xl backdrop-blur-md shadow-2xl">
        <div className="flex items-center gap-3">
          <ImageIcon className="w-5 h-5 text-indigo-400 shrink-0" />
          <h3 className="text-xs sm:text-sm font-bold text-white truncate max-w-xs sm:max-w-md">
            {title}
          </h3>
        </div>

        <div className="flex items-center gap-2">
          {/* Zoom controls */}
          <div className="flex items-center gap-1 bg-slate-800 p-1 rounded-xl border border-slate-700">
            <button
              type="button"
              onClick={handleZoomOut}
              className="p-1.5 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
              title="Zoom Out"
            >
              <ZoomOut className="w-4 h-4" />
            </button>
            <span className="text-xs font-mono font-bold text-indigo-300 px-2 min-w-[44px] text-center">
              {Math.round(scale * 100)}%
            </span>
            <button
              type="button"
              onClick={handleZoomIn}
              className="p-1.5 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
              title="Zoom In"
            >
              <ZoomIn className="w-4 h-4" />
            </button>
            <button
              type="button"
              onClick={handleReset}
              className="p-1.5 text-slate-400 hover:text-white hover:bg-slate-700 rounded-lg transition-colors ml-1 border-l border-slate-700"
              title="Reset Zoom"
            >
              <RotateCcw className="w-3.5 h-3.5" />
            </button>
          </div>

          <button
            type="button"
            onClick={onClose}
            className="p-2 text-slate-400 hover:text-white bg-slate-800 hover:bg-rose-600/30 rounded-xl border border-slate-700 transition-colors"
            title="Close Preview (Esc)"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Main Image Canvas Stage */}
      <div
        className="flex-1 flex items-center justify-center overflow-hidden relative my-4 cursor-grab active:cursor-grabbing"
        onWheel={handleWheel}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
      >
        <div
          style={{
            transform: `translate(${position.x}px, ${position.y}px) scale(${scale})`,
            transition: isDragging ? "none" : "transform 0.15s ease-out",
          }}
          className="max-w-full max-h-full flex items-center justify-center"
        >
          <img
            src={url}
            alt={title}
            draggable={false}
            className="max-h-[75vh] max-w-[85vw] object-contain rounded-2xl shadow-2xl drop-shadow-2xl border border-slate-800"
          />
        </div>
      </div>

      {/* Footer Instructions */}
      <div className="flex items-center justify-center gap-3 text-[11px] font-semibold text-slate-400 z-20 bg-slate-900/80 py-2 px-4 rounded-xl border border-slate-800 max-w-fit mx-auto shadow-lg">
        <span>💡 Scroll mouse wheel or pinch to zoom</span>
        <span>•</span>
        <span>Drag to pan when zoomed in</span>
      </div>
    </div>
  );
};

export const Step4OrderRequirement: React.FC<Step4OrderRequirementProps> = ({
  orderChoice,
  setOrderChoice,
  regularProducts,
  sampleProducts,
  orderQuantities,
  handleQuantityChange,
  calculateOrderTotal,
  onOpenOrderReview,
  onInitiateSampleBooking,
}) => {
  const [expandedCards, setExpandedCards] = useState<Record<string, boolean>>({});
  const [activeImageModal, setActiveImageModal] = useState<{ url: string; title: string } | null>(null);

  const openImageModal = (url: string, title: string, e?: React.MouseEvent) => {
    if (e) e.stopPropagation();
    setActiveImageModal({ url, title });
  };

  const toggleExpand = (id: string, e?: React.MouseEvent) => {
    if (e) e.stopPropagation();
    setExpandedCards((prev) => ({ ...prev, [id]: !prev[id] }));
  };
  // Filter out any sample products so full order catalog only contains regular commercial products
  const catalogProducts = regularProducts.filter(
    (p) => !p.is_sample && p.is_regular !== false
  );

  // Check if any selected product violates MOQ
  const hasMoqViolation = catalogProducts.some(
    (p) => (orderQuantities[p.id] || 0) > 0 && orderQuantities[p.id] < p.moq
  );

  const totalItemsCount = catalogProducts.reduce(
    (acc, p) => acc + (orderQuantities[p.id] || 0),
    0
  );

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 bg-indigo-500/10 text-indigo-400 rounded-xl flex items-center justify-center border border-indigo-500/20">
            <ShoppingCart className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-white">Step 4: Order Requirement Selection</h2>
            <p className="text-xs text-slate-400">Choose to place a catalog inventory order or book a sample kit</p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 my-6">
          {/* YES PATH CARD */}
          <div
            onClick={() => setOrderChoice("full")}
            className={`p-6 rounded-2xl border cursor-pointer transition-all ${
              orderChoice === "full"
                ? "bg-indigo-600/20 border-indigo-500 text-white ring-2 ring-indigo-500/30"
                : "bg-slate-800/40 border-slate-800 text-slate-400 hover:border-slate-700"
            }`}
          >
            <div className="flex items-center justify-between mb-3">
              <span className="px-3 py-1 bg-indigo-500/20 text-indigo-300 rounded-full text-xs font-bold uppercase tracking-wider">
                Option A: Full Order
              </span>
              {orderChoice === "full" && <CheckCircle2 className="w-5 h-5 text-indigo-400" />}
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Place Catalog Order Now</h3>
            <p className="text-xs text-slate-300 leading-relaxed mb-4">
              Select items from our active product catalog. Immediate full order unlocks priority credit scoring and statutory verification.
            </p>
            <div className="text-xs text-indigo-400 font-semibold flex items-center gap-1">
              Browse Catalogue ({catalogProducts.length} items) <ChevronRight className="w-4 h-4" />
            </div>
          </div>

          {/* NO PATH CARD */}
          <div
            onClick={() => setOrderChoice("sample")}
            className={`p-6 rounded-2xl border cursor-pointer transition-all ${
              orderChoice === "sample"
                ? "bg-amber-500/20 border-amber-500 text-white ring-2 ring-amber-500/30"
                : "bg-slate-800/40 border-slate-800 text-slate-400 hover:border-slate-700"
            }`}
          >
            <div className="flex items-center justify-between mb-3">
              <span className="px-3 py-1 bg-amber-500/20 text-amber-300 rounded-full text-xs font-bold uppercase tracking-wider flex items-center gap-1">
                <Sparkles className="w-3 h-3" /> Option B: Book Sample Kit
              </span>
              {orderChoice === "sample" && <CheckCircle2 className="w-5 h-5 text-amber-400" />}
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Order Trial Sample Kit</h3>
            <p className="text-xs text-slate-300 leading-relaxed mb-4">
              Test product quality first. Pay nominal sample charge via Razorpay to activate Instant Trial Status and bypass remaining credit steps.
            </p>
            <div className="text-xs text-amber-400 font-semibold flex items-center gap-1">
              Book Trial Kit <ChevronRight className="w-4 h-4" />
            </div>
          </div>
        </div>
      </div>

      {/* Render Catalog if Option A selected */}
      {orderChoice === "full" && (
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl">
          <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
            <Package className="w-5 h-5 text-indigo-400" />
            Regular Products Catalog
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            {catalogProducts.map((p) => {
              const isExpanded = !!expandedCards[p.id];
              const qty = orderQuantities[p.id] || 0;
              const isSubMoq = qty > 0 && qty < p.moq;

              const onPlus = (e: React.MouseEvent) => {
                e.stopPropagation();
                if (qty === 0) {
                  // Jump straight to MOQ when starting from 0
                  handleQuantityChange(p.id, p.moq);
                } else {
                  handleQuantityChange(p.id, 1);
                }
              };

              const onMinus = (e: React.MouseEvent) => {
                e.stopPropagation();
                if (qty <= p.moq) {
                  // Reset to 0 when decrementing at or below MOQ
                  handleQuantityChange(p.id, -qty);
                } else {
                  handleQuantityChange(p.id, -1);
                }
              };

              return (
                <div
                  key={p.id}
                  className={`border rounded-2xl overflow-hidden transition-all duration-300 ${
                    isSubMoq
                      ? "bg-rose-950/20 border-rose-500/50"
                      : qty > 0
                      ? "bg-slate-800/80 border-indigo-500/50"
                      : "bg-slate-800/40 border-slate-700/60 hover:border-slate-600"
                  }`}
                >
                  {/* Card Header */}
                  <div className="p-4 flex items-start justify-between gap-3">
                    <div
                      onClick={(e) => toggleExpand(p.id, e)}
                      className="flex items-start gap-3 flex-1 cursor-pointer group"
                    >
                      {p.image_url ? (
                        <div
                          onClick={(e) => openImageModal(p.image_url!, p.name, e)}
                          className="relative group/img cursor-zoom-in shrink-0"
                          title="Click to view full screen & zoom"
                        >
                          <img
                            src={p.image_url}
                            alt={p.name}
                            className="w-12 h-12 object-cover rounded-xl border border-slate-700 bg-slate-900 group-hover/img:scale-105 transition-transform"
                          />
                          <div className="absolute inset-0 bg-indigo-950/40 rounded-xl opacity-0 group-hover/img:opacity-100 flex items-center justify-center transition-opacity">
                            <ZoomIn className="w-4 h-4 text-white" />
                          </div>
                        </div>
                      ) : (
                        <div className="w-12 h-12 rounded-xl border border-slate-800 bg-slate-900 flex items-center justify-center text-slate-500 shrink-0 group-hover:border-indigo-500/40 transition-colors">
                          <Package className="w-6 h-6 text-indigo-400/60" />
                        </div>
                      )}
                      <div>
                        <div className="font-semibold text-white text-sm">
                          {p.name}
                        </div>
                        <div className="text-xs text-slate-400 mt-0.5">
                          SKU: {p.sku} |{" "}
                          <span className="font-semibold text-indigo-300">MOQ: {p.moq} units</span>
                        </div>
                        <div className="text-sm font-bold text-emerald-400 mt-1">
                          ₹{(p.price_paise / 100).toLocaleString("en-IN")}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 shrink-0">
                      <div className="flex items-center gap-2 bg-slate-900 px-3 py-1.5 rounded-xl border border-slate-700">
                        <button
                          type="button"
                          onClick={onMinus}
                          className="w-6 h-6 bg-slate-800 text-slate-300 rounded hover:bg-slate-700 font-bold flex items-center justify-center transition-colors"
                        >
                          -
                        </button>
                        <span className="w-8 text-center text-sm font-semibold">{qty}</span>
                        <button
                          type="button"
                          onClick={onPlus}
                          className="w-6 h-6 bg-indigo-600 text-white rounded hover:bg-indigo-500 font-bold flex items-center justify-center transition-colors"
                        >
                          +
                        </button>
                      </div>

                      <button
                        type="button"
                        onClick={(e) => toggleExpand(p.id, e)}
                        className="p-1.5 text-slate-400 hover:text-white rounded-lg hover:bg-slate-800 transition-colors"
                        title={isExpanded ? "Collapse details" : "Expand details"}
                      >
                        {isExpanded ? <ChevronUp className="w-5 h-5 text-indigo-400" /> : <ChevronDown className="w-5 h-5" />}
                      </button>
                    </div>
                  </div>

                  {/* Expanded Content Body */}
                  {isExpanded && (
                    <div className="px-4 pb-4 pt-2 border-t border-slate-700/40 bg-slate-950/40 space-y-3 animate-fadeIn">
                      {p.image_url && (
                        <div
                          onClick={(e) => openImageModal(p.image_url!, p.name, e)}
                          className="w-full h-48 rounded-xl overflow-hidden border border-slate-700/80 bg-slate-950 relative cursor-zoom-in group/expandimg"
                          title="Click to expand full screen & zoom"
                        >
                          <img
                            src={p.image_url}
                            alt={p.name}
                            className="w-full h-full object-contain p-2 group-hover/expandimg:scale-105 transition-transform"
                          />
                          <div className="absolute inset-0 bg-slate-950/40 opacity-0 group-hover/expandimg:opacity-100 flex items-center justify-center transition-opacity gap-2 text-xs font-semibold text-white">
                            <ZoomIn className="w-5 h-5 text-indigo-400" />
                            <span>Click to Zoom & Inspect</span>
                          </div>
                        </div>
                      )}

                      {p.description && (
                        <div>
                          <span className="text-[11px] font-bold uppercase text-slate-400 tracking-wider block mb-1">
                            Description & Specs
                          </span>
                          <p className="text-xs text-slate-300 leading-relaxed bg-slate-900/60 p-3 rounded-xl border border-slate-800">
                            {p.description}
                          </p>
                        </div>
                      )}

                      <div className="grid grid-cols-2 gap-2 text-xs pt-1">
                        <div className="bg-slate-900/80 p-2.5 rounded-xl border border-slate-800">
                          <span className="text-[10px] text-slate-400 block font-semibold">Category</span>
                          <span className="text-slate-200 font-bold">{p.category}</span>
                        </div>
                        <div className="bg-slate-900/80 p-2.5 rounded-xl border border-slate-800">
                          <span className="text-[10px] text-slate-400 block font-semibold">Min Order Qty</span>
                          <span className="text-indigo-300 font-bold">{p.moq} Units</span>
                        </div>
                      </div>
                    </div>
                  )}

                  {isSubMoq && (
                    <div className="m-4 mt-0 text-[11px] font-semibold text-rose-400 bg-rose-500/10 px-2.5 py-1 rounded-lg border border-rose-500/20">
                      ⚠️ Minimum order quantity for {p.name} is {p.moq}
                    </div>
                  )}
                </div>
              );
            })}
          </div>

          <div className="flex items-center justify-between pt-4 border-t border-slate-800">
            <div>
              <span className="text-xs text-slate-400">Total Order Value:</span>
              <div className="text-2xl font-extrabold text-emerald-400">
                ₹{calculateOrderTotal().toLocaleString("en-IN")}
              </div>
            </div>

            <button
              type="button"
              onClick={onOpenOrderReview}
              disabled={totalItemsCount === 0 || hasMoqViolation}
              className="px-6 py-3 bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-medium rounded-xl shadow-lg shadow-indigo-600/30 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              <span>Review & Submit Order</span>
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      {/* Render Sample Kits if Option B selected */}
      {orderChoice === "sample" && (
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl">
          <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-amber-400" />
            Available Trial Sample Kits
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            {sampleProducts.map((sp) => {
              const isExpanded = !!expandedCards[sp.id];

              return (
                <div
                  key={sp.id}
                  className="border border-amber-500/30 rounded-2xl overflow-hidden bg-gradient-to-br from-slate-800/80 to-slate-900 transition-all duration-300 flex flex-col justify-between"
                >
                  <div className="p-5 space-y-4">
                    <div
                      onClick={(e) => toggleExpand(sp.id, e)}
                      className="flex items-start gap-3 cursor-pointer group"
                    >
                      {sp.image_url ? (
                        <div
                          onClick={(e) => openImageModal(sp.image_url!, sp.name, e)}
                          className="relative group/sampleimg cursor-zoom-in shrink-0"
                          title="Click to view full screen & zoom"
                        >
                          <img
                            src={sp.image_url}
                            alt={sp.name}
                            className="w-14 h-14 object-cover rounded-xl border border-amber-500/30 bg-slate-900 group-hover/sampleimg:scale-105 transition-transform"
                          />
                          <div className="absolute inset-0 bg-amber-950/40 rounded-xl opacity-0 group-hover/sampleimg:opacity-100 flex items-center justify-center transition-opacity">
                            <ZoomIn className="w-4 h-4 text-amber-300" />
                          </div>
                        </div>
                      ) : (
                        <div className="w-14 h-14 rounded-xl border border-amber-500/20 bg-amber-500/10 flex items-center justify-center text-amber-400 shrink-0 group-hover:border-amber-400/40 transition-colors">
                          <Sparkles className="w-7 h-7" />
                        </div>
                      )}
                      <div className="flex-1">
                        <div className="flex items-center justify-between">
                          <span className="px-2.5 py-0.5 bg-amber-500/20 text-amber-300 rounded-md text-[11px] font-semibold uppercase">
                            Sample Trial Kit
                          </span>
                          <button
                            type="button"
                            onClick={(e) => toggleExpand(sp.id, e)}
                            className="text-slate-400 hover:text-amber-300 p-1 rounded-lg hover:bg-slate-800/80 transition-colors flex items-center gap-1 text-xs font-semibold"
                          >
                            <span>{isExpanded ? "Hide Details" : "View Image & Specs"}</span>
                            {isExpanded ? <ChevronUp className="w-4 h-4 text-amber-400" /> : <ChevronDown className="w-4 h-4" />}
                          </button>
                        </div>
                        <h4 className="font-bold text-white text-base mt-1.5 group-hover:text-amber-300 transition-colors">
                          {sp.name}
                        </h4>
                        <p className="text-xs text-slate-400 mt-1 line-clamp-2">
                          {sp.description || "Official trial bundle for distributor evaluation"}
                        </p>
                      </div>
                    </div>

                    {/* Expanded Section for Sample Kit */}
                    {isExpanded && (
                      <div className="pt-3 border-t border-amber-500/20 space-y-3 animate-fadeIn">
                        {sp.image_url && (
                          <div
                            onClick={(e) => openImageModal(sp.image_url!, sp.name, e)}
                            className="w-full h-48 rounded-xl overflow-hidden border border-amber-500/30 bg-slate-950 relative cursor-zoom-in group/expandsample"
                            title="Click to expand full screen & zoom"
                          >
                            <img
                              src={sp.image_url}
                              alt={sp.name}
                              className="w-full h-full object-contain p-2 group-hover/expandsample:scale-105 transition-transform"
                            />
                            <div className="absolute inset-0 bg-slate-950/40 opacity-0 group-hover/expandsample:opacity-100 flex items-center justify-center transition-opacity gap-2 text-xs font-semibold text-amber-300">
                              <ZoomIn className="w-5 h-5 text-amber-400" />
                              <span>Click to Zoom & Inspect</span>
                            </div>
                          </div>
                        )}

                        <div className="bg-slate-950/60 p-3 rounded-xl border border-slate-800 space-y-1">
                          <span className="text-[10px] font-bold uppercase text-amber-400 tracking-wider block">
                            Kit Inclusions & Evaluation Specs
                          </span>
                          <p className="text-xs text-slate-300 leading-relaxed">
                            {sp.description || "Contains curated product samples, brand collateral, and official distributor wholesale catalog pricing sheet."}
                          </p>
                        </div>
                      </div>
                    )}
                  </div>

                  <div className="p-5 pt-0 flex items-center justify-between">
                    <div>
                      <div className="text-xs text-slate-400">Sample Price</div>
                      <div className="text-xl font-bold text-amber-400">
                        ₹{(sp.price_paise / 100).toLocaleString("en-IN")}
                      </div>
                    </div>

                    <button
                      type="button"
                      onClick={() => onInitiateSampleBooking(sp)}
                      className="px-4 py-2.5 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-400 hover:to-orange-400 text-slate-950 font-bold text-xs rounded-xl shadow-lg shadow-amber-500/20 flex items-center gap-1.5 transition-transform active:scale-95"
                    >
                      <span>Book via Razorpay</span>
                      <ArrowRight className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Fullscreen Image Zoom Modal */}
      {activeImageModal && (
        <ImageZoomModal
          url={activeImageModal.url}
          title={activeImageModal.title}
          onClose={() => setActiveImageModal(null)}
        />
      )}
    </div>
  );
};
