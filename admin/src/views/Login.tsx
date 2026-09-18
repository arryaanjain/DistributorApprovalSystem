import React, { useState } from 'react';
import { Mail, Lock, ShieldCheck, AlertCircle } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { api } from '../services/api';

export const Login: React.FC = () => {
  const { login } = useAuth();
  const defaultEmail = import.meta.env.VITE_ADMIN_DEFAULT_EMAIL || 'kresconet@gmail.com';
  const defaultPassword = import.meta.env.VITE_ADMIN_DEFAULT_PASSWORD || 'Kresco4572?';

  const [email, setEmail] = useState(defaultEmail);
  const [password, setPassword] = useState(defaultPassword);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const data = await api.loginWithCredentials(email, password);
      const token = data.access_token || data.token;
      if (token) {
        login(token, data.user || { id: 'EMP-ADMIN', name: 'Kresconet Admin', role: 'super_admin' }, data.refresh_token);
      } else {
        setError('Authentication failed. No access token received.');
      }
    } catch (err: any) {
      setError(err.message || 'Invalid credentials or server unavailable.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-slate-900/90 border border-indigo-500/30 rounded-3xl p-8 shadow-2xl backdrop-blur-xl">
        <div className="text-center mb-8">
          <div className="w-16 h-16 bg-gradient-to-tr from-indigo-600 to-violet-500 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-lg shadow-indigo-500/30">
            <ShieldCheck className="w-9 h-9 text-white" />
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Kresconet Admin Console</h1>
          <p className="text-xs text-slate-400 mt-1">
            Sign in with authorized employee credentials
          </p>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-rose-500/10 border border-rose-500/30 rounded-xl text-rose-400 text-xs flex items-center gap-3">
            <AlertCircle className="w-5 h-5 flex-shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
              Email Address
            </label>
            <div className="relative">
              <Mail className="absolute left-3.5 top-3.5 w-5 h-5 text-slate-500" />
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                className="w-full pl-11 pr-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl focus:outline-none focus:border-indigo-500 text-slate-100 text-sm placeholder-slate-500"
                placeholder="admin@kresconet.com"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
              Password
            </label>
            <div className="relative">
              <Lock className="absolute left-3.5 top-3.5 w-5 h-5 text-slate-500" />
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="w-full pl-11 pr-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl focus:outline-none focus:border-indigo-500 text-slate-100 text-sm placeholder-slate-500"
                placeholder="••••••••••••"
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3.5 px-4 bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-semibold rounded-xl shadow-lg shadow-indigo-600/30 transition-all flex items-center justify-center gap-2"
          >
            {loading ? (
              <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
            ) : (
              <span>Sign In to Admin Portal</span>
            )}
          </button>

          <div className="mt-4 p-3 bg-indigo-950/40 border border-indigo-500/20 rounded-xl text-xs text-indigo-300 text-center">
            Seeded Admin Account: <br />
            <span className="font-mono text-white font-semibold">{defaultEmail} / {defaultPassword}</span>
          </div>
        </form>
      </div>
    </div>
  );
};
