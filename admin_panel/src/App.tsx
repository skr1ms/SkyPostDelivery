import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './context/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import Layout from './components/Layout/Layout';
import LoginPage from './pages/Login/LoginPage';
import Dashboard from './pages/Dashboard/Dashboard';
import DronesPage from './pages/Drones/DronesPage';
import DroneDetailPage from './pages/Drones/DroneDetailPage';
import GoodsPage from './pages/Goods/GoodsPage';
import ParcelAutomatsPage from './pages/ParcelAutomats/ParcelAutomatsPage';
import MonitoringPage from './pages/Monitoring/MonitoringPage';

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="drones" element={<DronesPage />} />
            <Route path="drones/:droneId" element={<DroneDetailPage />} />
            <Route path="goods" element={<GoodsPage />} />
            <Route path="parcel-automats" element={<ParcelAutomatsPage />} />
            <Route path="monitoring" element={<MonitoringPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
      <Toaster />
    </AuthProvider>
  );
}

export default App;
