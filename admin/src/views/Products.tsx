import React, { useState, useEffect } from 'react';
import { 
  Package, 
  Plus, 
  Edit3, 
  CheckCircle2, 
  XCircle, 
  Sparkles, 
  Search, 
  AlertCircle,
  ShoppingBag,
  Layers
} from 'lucide-react';
import { api } from '../services/api';

interface Product {
  id?: string;
  sku: string;
  name: string;
  description?: string;
  category: string;
  price_paise: number;
  moq: number;
  is_active: boolean;
  is_sample: boolean;
  is_regular: boolean;
}

type TabType = 'regular' | 'sample' | 'all';

export const Products: React.FC = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [activeTab, setActiveTab] = useState<TabType>('regular');
  const [showModal, setShowModal] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState<Product>({
    sku: '',
    name: '',
    description: '',
    category: 'Edible Oils',
    price_paise: 50000,
    moq: 1,
    is_active: true,
    is_sample: false,
    is_regular: true,
  });

  useEffect(() => {
    loadProducts();
  }, []);

  const loadProducts = async () => {
    setLoading(true);
    try {
      const data = await api.listProductsAdmin();
      setProducts(Array.isArray(data) ? data : []);
    } catch (err: any) {
      console.error('Failed to load products', err);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenCreate = (targetScope: 'regular' | 'sample' = 'regular') => {
    setEditingProduct(null);
    const isSample = targetScope === 'sample';
    setFormData({
      sku: isSample ? `SMP-${Math.floor(100 + Math.random() * 900)}` : `PROD-${Math.floor(100 + Math.random() * 900)}`,
      name: '',
      description: '',
      category: isSample ? 'Sample Pack' : 'Edible Oils',
      price_paise: isSample ? 29900 : 50000,
      moq: 1,
      is_active: true,
      is_sample: isSample,
      is_regular: !isSample,
    });
    setError(null);
    setShowModal(true);
  };

  const handleOpenEdit = (p: Product) => {
    setEditingProduct(p);
    setFormData({ ...p });
    setError(null);
    setShowModal(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!formData.sku.trim() || !formData.name.trim()) {
      setError('SKU and Product Name are required.');
      return;
    }

    if (!formData.is_regular && !formData.is_sample) {
      setError('Please select at least one order scope (Regular Order or Sample Trial).');
      return;
    }

    try {
      if (editingProduct?.id) {
        await api.updateProduct(editingProduct.id, formData);
      } else {
        await api.createProduct(formData);
      }
      setShowModal(false);
      loadProducts();
    } catch (err: any) {
      setError(err.message || 'Failed to save product');
    }
  };

  const regularProducts = products.filter((p) => p.is_regular);
  const sampleProducts = products.filter((p) => p.is_sample);

  const displayedProducts = products.filter((p) => {
    const matchesSearch =
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.sku.toLowerCase().includes(search.toLowerCase()) ||
      p.category.toLowerCase().includes(search.toLowerCase());

    if (!matchesSearch) return false;

    if (activeTab === 'regular') return p.is_regular;
    if (activeTab === 'sample') return p.is_sample;
    return true;
  });

  return (
    <div className="p-8 space-y-6 max-w-7xl mx-auto">
      {/* Header Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight flex items-center gap-3">
            <Package className="w-7 h-7 text-indigo-400" />
            Bifurcated Product Catalogue Management
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Manage distributor commercial purchasing products and onboarding sample trial kits separately.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <div className="relative">
            <Search className="absolute left-3.5 top-2.5 w-4 h-4 text-slate-500" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search by SKU, name, category..."
              className="pl-10 pr-4 py-2 bg-slate-900 border border-slate-800 rounded-xl text-xs text-slate-200 focus:outline-none focus:border-indigo-500 w-64"
            />
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => handleOpenCreate('regular')}
              className="px-3.5 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-xl shadow-lg shadow-indigo-600/30 transition-all flex items-center gap-1.5"
            >
              <Plus className="w-4 h-4" />
              <span>Add Commercial Product</span>
            </button>

            <button
              onClick={() => handleOpenCreate('sample')}
              className="px-3.5 py-2 bg-amber-600 hover:bg-amber-500 text-white text-xs font-bold rounded-xl shadow-lg shadow-amber-600/30 transition-all flex items-center gap-1.5"
            >
              <Sparkles className="w-4 h-4" />
              <span>Add Sample Trial Kit</span>
            </button>
          </div>
        </div>
      </div>

      {/* Metrics Bifurcation Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div 
          onClick={() => setActiveTab('regular')}
          className={`glass-panel p-5 rounded-2xl border transition-all cursor-pointer ${
            activeTab === 'regular' 
              ? 'border-indigo-500/60 bg-indigo-950/20 shadow-lg shadow-indigo-500/10' 
              : 'border-slate-800 hover:border-slate-700'
          }`}
        >
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
              <ShoppingBag className="w-4 h-4 text-indigo-400" /> Commercial Products
            </span>
            <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
              Regular Orders
            </span>
          </div>
          <p className="text-3xl font-black text-white mt-3">{regularProducts.length}</p>
          <p className="text-xs text-slate-400 mt-1">Available for distributor credit line ordering</p>
        </div>

        <div 
          onClick={() => setActiveTab('sample')}
          className={`glass-panel p-5 rounded-2xl border transition-all cursor-pointer ${
            activeTab === 'sample' 
              ? 'border-amber-500/60 bg-amber-950/20 shadow-lg shadow-amber-500/10' 
              : 'border-slate-800 hover:border-slate-700'
          }`}
        >
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
              <Sparkles className="w-4 h-4 text-amber-400" /> Sample Trial Kits
            </span>
            <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30">
              Prepaid Trial
            </span>
          </div>
          <p className="text-3xl font-black text-amber-400 mt-3">{sampleProducts.length}</p>
          <p className="text-xs text-slate-400 mt-1">Available during Step 4 onboarding checkout</p>
        </div>

        <div 
          onClick={() => setActiveTab('all')}
          className={`glass-panel p-5 rounded-2xl border transition-all cursor-pointer ${
            activeTab === 'all' 
              ? 'border-violet-500/60 bg-violet-950/20 shadow-lg shadow-violet-500/10' 
              : 'border-slate-800 hover:border-slate-700'
          }`}
        >
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
              <Layers className="w-4 h-4 text-violet-400" /> Complete Inventory
            </span>
            <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-slate-800 text-slate-300">
              All Catalogue
            </span>
          </div>
          <p className="text-3xl font-black text-white mt-3">{products.length}</p>
          <p className="text-xs text-slate-400 mt-1">Total active & inactive catalog items</p>
        </div>
      </div>

      {/* Tabs Bar */}
      <div className="flex items-center gap-2 border-b border-slate-800 pb-3">
        <button
          onClick={() => setActiveTab('regular')}
          className={`px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2 ${
            activeTab === 'regular'
              ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/30'
              : 'bg-slate-900/60 text-slate-400 hover:text-slate-200 hover:bg-slate-800/60 border border-slate-800'
          }`}
        >
          <ShoppingBag className="w-4 h-4" />
          Commercial Products ({regularProducts.length})
        </button>

        <button
          onClick={() => setActiveTab('sample')}
          className={`px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2 ${
            activeTab === 'sample'
              ? 'bg-amber-600 text-white shadow-lg shadow-amber-600/30'
              : 'bg-slate-900/60 text-slate-400 hover:text-slate-200 hover:bg-slate-800/60 border border-slate-800'
          }`}
        >
          <Sparkles className="w-4 h-4 text-amber-300" />
          Sample Trial Kits ({sampleProducts.length})
        </button>

        <button
          onClick={() => setActiveTab('all')}
          className={`px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2 ${
            activeTab === 'all'
              ? 'bg-slate-700 text-white shadow-md'
              : 'bg-slate-900/60 text-slate-400 hover:text-slate-200 hover:bg-slate-800/60 border border-slate-800'
          }`}
        >
          <Layers className="w-4 h-4" />
          All Catalog Items ({products.length})
        </button>
      </div>

      {/* Table */}
      {loading ? (
        <div className="py-20 text-center text-slate-500 text-sm">Loading catalogue...</div>
      ) : displayedProducts.length === 0 ? (
        <div className="py-20 text-center bg-slate-900/40 border border-slate-800/80 rounded-2xl space-y-3">
          <Package className="w-12 h-12 text-slate-600 mx-auto" />
          <p className="text-slate-400 font-semibold text-sm">No products found in this view</p>
          <div className="flex justify-center gap-3">
            <button
              onClick={() => handleOpenCreate(activeTab === 'sample' ? 'sample' : 'regular')}
              className="px-4 py-2 bg-indigo-600 text-white text-xs font-semibold rounded-xl"
            >
              Create Product Now
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-slate-900/60 border border-slate-800 rounded-2xl overflow-hidden shadow-xl">
          <table className="w-full text-left text-sm text-slate-300">
            <thead className="bg-slate-900/90 text-[11px] uppercase font-bold text-slate-400 border-b border-slate-800">
              <tr>
                <th className="px-6 py-4">SKU / Product Name</th>
                <th className="px-6 py-4">Category</th>
                <th className="px-6 py-4">Price (₹)</th>
                <th className="px-6 py-4">MOQ</th>
                <th className="px-6 py-4">Bifurcated Scope</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {displayedProducts.map((p) => (
                <tr key={p.id || p.sku} className="hover:bg-slate-800/30 transition-colors">
                  <td className="px-6 py-4">
                    <div className="font-bold text-white flex items-center gap-2">
                      <span>{p.name}</span>
                      {p.is_sample && (
                        <span className="px-2 py-0.5 rounded text-[10px] font-extrabold bg-amber-500/20 text-amber-300 border border-amber-500/30 uppercase">
                          Sample Kit
                        </span>
                      )}
                    </div>
                    <div className="text-xs text-slate-500 font-mono mt-0.5">{p.sku}</div>
                    {p.description && <p className="text-[11px] text-slate-400 mt-1 line-clamp-1">{p.description}</p>}
                  </td>
                  <td className="px-6 py-4">
                    <span className="px-2.5 py-1 bg-slate-800 text-slate-300 rounded-lg text-xs font-medium border border-slate-700/60">
                      {p.category}
                    </span>
                  </td>
                  <td className="px-6 py-4 font-bold text-emerald-400">
                    ₹{(p.price_paise / 100).toLocaleString('en-IN')}
                  </td>
                  <td className="px-6 py-4 text-slate-300">{p.moq} units</td>
                  <td className="px-6 py-4">
                    <div className="flex gap-1.5 flex-wrap">
                      {p.is_regular && (
                        <span className="px-2 py-0.5 bg-indigo-500/10 text-indigo-300 border border-indigo-500/30 rounded-md text-[11px] font-semibold flex items-center gap-1">
                          <ShoppingBag className="w-3 h-3 text-indigo-400" /> Commercial Purchase
                        </span>
                      )}
                      {p.is_sample && (
                        <span className="px-2 py-0.5 bg-amber-500/10 text-amber-300 border border-amber-500/30 rounded-md text-[11px] font-semibold flex items-center gap-1">
                          <Sparkles className="w-3 h-3 text-amber-400" /> Trial Booking
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    {p.is_active ? (
                      <span className="inline-flex items-center gap-1 text-xs text-emerald-400 font-semibold">
                        <CheckCircle2 className="w-3.5 h-3.5" /> Active
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 text-xs text-slate-500 font-semibold">
                        <XCircle className="w-3.5 h-3.5" /> Inactive
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button
                      onClick={() => handleOpenEdit(p)}
                      className="p-2 text-indigo-400 hover:bg-indigo-500/10 rounded-lg transition-colors"
                      title="Edit Product"
                    >
                      <Edit3 className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Bifurcated Product Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="w-full max-w-lg bg-slate-900 border border-slate-700 rounded-3xl p-6 shadow-2xl max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4 border-b border-slate-800 pb-3">
              <h3 className="text-lg font-bold text-white flex items-center gap-2">
                {formData.is_sample ? (
                  <>
                    <Sparkles className="w-5 h-5 text-amber-400" />
                    <span>{editingProduct ? 'Edit Sample Trial Kit' : 'Add Sample Trial Kit'}</span>
                  </>
                ) : (
                  <>
                    <ShoppingBag className="w-5 h-5 text-indigo-400" />
                    <span>{editingProduct ? 'Edit Commercial Product' : 'Add Commercial Product'}</span>
                  </>
                )}
              </h3>
              <button
                onClick={() => setShowModal(false)}
                className="text-slate-400 hover:text-white text-sm"
              >
                ✕
              </button>
            </div>

            {error && (
              <div className="mb-4 p-3 bg-rose-500/10 border border-rose-500/30 rounded-xl text-rose-400 text-xs flex items-center gap-2">
                <AlertCircle className="w-4 h-4 flex-shrink-0" />
                <span>{error}</span>
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">SKU</label>
                  <input
                    type="text"
                    value={formData.sku}
                    onChange={(e) => setFormData({ ...formData, sku: e.target.value })}
                    required
                    disabled={!!editingProduct}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500 disabled:opacity-60 font-mono"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">Category</label>
                  <input
                    type="text"
                    value={formData.category}
                    onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                    required
                    placeholder="e.g. Edible Oils, Staples, Sample Pack"
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">Product Title</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                  placeholder={formData.is_sample ? "e.g. Sample Pack: Premium Staples & Grains" : "e.g. Kresco Gold Refined Sunflower Oil (15L Tin)"}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">Description</label>
                <textarea
                  value={formData.description || ''}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows={2}
                  placeholder="Detailed specs or trial kit breakdown..."
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500 text-xs"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">Price (₹ INR)</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.price_paise / 100}
                    onChange={(e) => setFormData({ ...formData, price_paise: Math.round(parseFloat(e.target.value || '0') * 100) })}
                    required
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">Minimum Order Qty (MOQ)</label>
                  <input
                    type="number"
                    value={formData.moq}
                    onChange={(e) => setFormData({ ...formData, moq: parseInt(e.target.value || '1', 10) })}
                    required
                    min={1}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              {/* Bifurcated Scope Checkboxes */}
              <div className="p-4 bg-slate-800/60 border border-slate-700/60 rounded-xl space-y-3">
                <label className="block text-xs font-bold uppercase text-indigo-400 mb-2">Bifurcated Order Scope Assignment</label>

                <div className="flex items-center gap-3">
                  <input
                    type="checkbox"
                    id="is_regular"
                    checked={formData.is_regular}
                    onChange={(e) => setFormData({ ...formData, is_regular: e.target.checked })}
                    className="w-4 h-4 text-indigo-600 rounded bg-slate-900 border-slate-700"
                  />
                  <label htmlFor="is_regular" className="text-xs font-medium text-slate-200">
                    Commercial Catalogue Product (Available for credit line bulk ordering)
                  </label>
                </div>

                <div className="flex items-center gap-3">
                  <input
                    type="checkbox"
                    id="is_sample"
                    checked={formData.is_sample}
                    onChange={(e) => setFormData({ ...formData, is_sample: e.target.checked })}
                    className="w-4 h-4 text-amber-500 rounded bg-slate-900 border-slate-700"
                  />
                  <label htmlFor="is_sample" className="text-xs font-medium text-slate-200 flex items-center gap-1.5">
                    <Sparkles className="w-3.5 h-3.5 text-amber-400" />
                    Onboarding Sample Trial Kit (Available for Step 4 prepaid checkout)
                  </label>
                </div>

                <div className="flex items-center gap-3 pt-2 border-t border-slate-700/40">
                  <input
                    type="checkbox"
                    id="is_active"
                    checked={formData.is_active}
                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                    className="w-4 h-4 text-emerald-500 rounded bg-slate-900 border-slate-700"
                  />
                  <label htmlFor="is_active" className="text-xs font-medium text-slate-200">
                    Active (Visible to Distributors)
                  </label>
                </div>
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t border-slate-800">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="px-4 py-2 bg-slate-800 text-slate-300 text-xs font-semibold rounded-xl"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-xl shadow-lg shadow-indigo-600/30"
                >
                  {editingProduct ? 'Update Catalogue Item' : 'Create Catalogue Item'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
