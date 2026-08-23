import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { Sidebar } from './components/Sidebar';
import { Navbar } from './components/Navbar';
import { Dashboard } from './views/Dashboard';
import { Applications } from './views/Applications';
import { Distributors } from './views/Distributors';
import { Orders } from './views/Orders';
import { Policy } from './views/Policy';

const AdminLayout: React.FC = () => {
  return (
    <div className="flex min-h-screen bg-slate-950 text-slate-100">
      <Sidebar />
      <div className="flex-1 flex flex-col min-w-0">
        <Navbar />
        <main className="flex-1 overflow-y-auto">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/applications" element={<Applications />} />
            <Route path="/distributors" element={<Distributors />} />
            <Route path="/orders" element={<Orders />} />
            <Route path="/policy" element={<Policy />} />
          </Routes>
        </main>
      </div>
    </div>
  );
};

export const App: React.FC = () => {
  return (
    <AuthProvider>
      <BrowserRouter>
        <AdminLayout />
      </BrowserRouter>
    </AuthProvider>
  );
};

export default App;
